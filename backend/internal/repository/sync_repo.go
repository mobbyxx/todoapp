package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type syncRepository struct {
	db *pgxpool.Pool
}

func NewSyncRepository(db *pgxpool.Pool) domain.SyncRepository {
	return &syncRepository{db: db}
}

func (r *syncRepository) GetLastSync(userID uuid.UUID) (*domain.SyncRecord, error) {
	query := `
		SELECT id, user_id, last_synced_at_physical, last_synced_at_logical, 
		       status, error_message, client_version, created_at, updated_at
		FROM sync_records
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var record domain.SyncRecord
	var lastPhysical, lastLogical int64
	var errorMsg *string

	err := r.db.QueryRow(context.Background(), query, userID).Scan(
		&record.ID,
		&record.UserID,
		&lastPhysical,
		&lastLogical,
		&record.Status,
		&errorMsg,
		&record.ClientVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	record.LastSyncedAt = domain.HLC{
		PhysicalTime: lastPhysical,
		LogicalTime:  lastLogical,
	}
	record.ErrorMessage = errorMsg

	return &record, nil
}

func (r *syncRepository) UpdateLastSync(userID uuid.UUID, timestamp domain.HLC) error {
	query := `
		INSERT INTO sync_records (user_id, last_synced_at_physical, last_synced_at_logical, 
		                          status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			last_synced_at_physical = EXCLUDED.last_synced_at_physical,
			last_synced_at_logical = EXCLUDED.last_synced_at_logical,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := r.db.Exec(
		context.Background(),
		query,
		userID,
		timestamp.PhysicalTime,
		timestamp.LogicalTime,
		domain.SyncStatusCompleted,
		now,
		now,
	)

	return err
}

func (r *syncRepository) GetChangesSince(userID uuid.UUID, since domain.HLC) (*domain.ChangeSet, error) {
	changes := &domain.ChangeSet{
		Created: []*domain.TodoChange{},
		Updated: []*domain.TodoChange{},
		Deleted: []*domain.TodoChange{},
	}

	query := `
		SELECT id, title, description, status, priority, created_by, assigned_to, 
		       due_date, version, created_at, updated_at, deleted_at,
		       EXTRACT(EPOCH FROM created_at)::bigint * 1000 as created_ms,
		       EXTRACT(EPOCH FROM updated_at)::bigint * 1000 as updated_ms,
		       EXTRACT(EPOCH FROM deleted_at)::bigint * 1000 as deleted_ms
		FROM todos
		WHERE (created_by = $1 OR assigned_to = $1)
		  AND (
			  (EXTRACT(EPOCH FROM created_at)::bigint * 1000 > $2) OR
			  (EXTRACT(EPOCH FROM updated_at)::bigint * 1000 > $2)
		  )
		ORDER BY updated_at ASC
	`

	rows, err := r.db.Query(context.Background(), query, userID, since.PhysicalTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo domain.Todo
		var assignedTo *uuid.UUID
		var dueDate, deletedAt *time.Time
		var createdMs, updatedMs int64
		var deletedMs *int64

		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Status,
			&todo.Priority,
			&todo.CreatedBy,
			&assignedTo,
			&dueDate,
			&todo.Version,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&deletedAt,
			&createdMs,
			&updatedMs,
			&deletedMs,
		)
		if err != nil {
			return nil, err
		}

		if assignedTo != nil {
			todo.AssignedTo = assignedTo
		}
		if dueDate != nil {
			todo.DueDate = dueDate
		}
		if deletedAt != nil {
			todo.DeletedAt = deletedAt
		}

		change := domain.TodoChangeFromTodo(&todo)

		if deletedAt != nil {
			change.IsDeleted = true
			change.Timestamp = domain.HLC{
				PhysicalTime: *deletedMs,
				LogicalTime:  0,
			}
			changes.Deleted = append(changes.Deleted, change)
		} else if createdMs > since.PhysicalTime && createdMs == updatedMs {
			change.Timestamp = domain.HLC{
				PhysicalTime: createdMs,
				LogicalTime:  0,
			}
			changes.Created = append(changes.Created, change)
		} else {
			change.Timestamp = domain.HLC{
				PhysicalTime: updatedMs,
				LogicalTime:  0,
			}
			changes.Updated = append(changes.Updated, change)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return changes, nil
}

func (r *syncRepository) ApplyChanges(userID uuid.UUID, changes domain.ChangeSet, timestamp domain.HLC) ([]*domain.SyncConflict, error) {
	conflicts := []*domain.SyncConflict{}
	now := time.Now()

	err := r.withTx(context.Background(), func(tx pgx.Tx) error {
		for _, created := range changes.Created {
			if err := r.applyCreate(tx, userID, created, now); err != nil {
				if errors.Is(err, domain.ErrConflictDetected) {
					conflict, err := r.handleConflict(tx, userID, created, nil, "create")
					if err != nil {
						return err
					}
					conflicts = append(conflicts, conflict)
				} else {
					return err
				}
			}
		}

		for _, updated := range changes.Updated {
			if err := r.applyUpdate(tx, userID, updated, now); err != nil {
				if errors.Is(err, domain.ErrConflictDetected) {
					conflict, err := r.handleConflict(tx, userID, updated, nil, "update")
					if err != nil {
						return err
					}
					conflicts = append(conflicts, conflict)
				} else {
					return err
				}
			}
		}

		for _, deleted := range changes.Deleted {
			if err := r.applyDelete(tx, userID, deleted, now); err != nil {
				if errors.Is(err, domain.ErrConflictDetected) {
					conflict, err := r.handleConflict(tx, userID, deleted, nil, "delete")
					if err != nil {
						return err
					}
					conflicts = append(conflicts, conflict)
				} else {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return conflicts, nil
}

func (r *syncRepository) applyCreate(tx pgx.Tx, userID uuid.UUID, change *domain.TodoChange, now time.Time) error {
	query := `
		INSERT INTO todos (id, title, description, status, priority, created_by, assigned_to, 
		                   due_date, version, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			assigned_to = EXCLUDED.assigned_to,
			due_date = EXCLUDED.due_date,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
		WHERE todos.version < EXCLUDED.version
			OR (todos.version = EXCLUDED.version 
			    AND EXTRACT(EPOCH FROM todos.updated_at)::bigint * 1000 < $13)
	`

	_, err := tx.Exec(
		context.Background(),
		query,
		change.ID,
		change.Title,
		change.Description,
		change.Status,
		change.Priority,
		change.CreatedBy,
		change.AssignedTo,
		change.DueDate,
		change.Version,
		now,
		now,
		nil,
		change.Timestamp.PhysicalTime,
	)

	return err
}

func (r *syncRepository) applyUpdate(tx pgx.Tx, userID uuid.UUID, change *domain.TodoChange, now time.Time) error {
	query := `
		UPDATE todos
		SET title = $1, description = $2, status = $3, priority = $4, assigned_to = $5,
		    due_date = $6, version = $7, updated_at = $8
		WHERE id = $9 
		  AND (created_by = $10 OR assigned_to = $10)
		  AND (version < $7 
		       OR (version = $7 
		           AND EXTRACT(EPOCH FROM updated_at)::bigint * 1000 < $11))
	`

	result, err := tx.Exec(
		context.Background(),
		query,
		change.Title,
		change.Description,
		change.Status,
		change.Priority,
		change.AssignedTo,
		change.DueDate,
		change.Version,
		now,
		change.ID,
		userID,
		change.Timestamp.PhysicalTime,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrConflictDetected
	}

	return nil
}

func (r *syncRepository) applyDelete(tx pgx.Tx, userID uuid.UUID, change *domain.TodoChange, now time.Time) error {
	query := `
		UPDATE todos
		SET deleted_at = $1, updated_at = $1, is_deleted = true
		WHERE id = $2 
		  AND (created_by = $3 OR assigned_to = $3)
		  AND deleted_at IS NULL
	`

	result, err := tx.Exec(
		context.Background(),
		query,
		now,
		change.ID,
		userID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		existing, err := r.getTodo(tx, change.ID)
		if err != nil {
			return err
		}
		if existing != nil && existing.DeletedAt == nil {
			return domain.ErrConflictDetected
		}
	}

	return nil
}

func (r *syncRepository) handleConflict(tx pgx.Tx, userID uuid.UUID, local *domain.TodoChange, remote *domain.Todo, changeType string) (*domain.SyncConflict, error) {
	serverTodo, err := r.getTodo(tx, local.ID)
	if err != nil {
		return nil, err
	}

	conflict := &domain.SyncConflict{
		EntityType:   "todo",
		EntityID:     local.ID,
		LocalVersion: local,
		ConflictType: changeType + "_conflict",
	}

	if serverTodo != nil {
		serverChange := domain.TodoChangeFromTodo(serverTodo)
		conflict.ServerVersion = serverChange
		conflict.ConflictType = "both_modified"
	}

	if err := r.recordConflict(tx, userID, conflict, local); err != nil {
		return nil, err
	}

	return conflict, nil
}

func (r *syncRepository) getTodo(tx pgx.Tx, id uuid.UUID) (*domain.Todo, error) {
	query := `
		SELECT id, title, description, status, priority, created_by, assigned_to, 
		       due_date, version, created_at, updated_at, deleted_at
		FROM todos
		WHERE id = $1
	`

	var todo domain.Todo
	var assignedTo *uuid.UUID
	var dueDate, deletedAt *time.Time

	err := tx.QueryRow(context.Background(), query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Status,
		&todo.Priority,
		&todo.CreatedBy,
		&assignedTo,
		&dueDate,
		&todo.Version,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if assignedTo != nil {
		todo.AssignedTo = assignedTo
	}
	if dueDate != nil {
		todo.DueDate = dueDate
	}
	if deletedAt != nil {
		todo.DeletedAt = deletedAt
	}

	return &todo, nil
}

func (r *syncRepository) recordConflict(tx pgx.Tx, userID uuid.UUID, conflict *domain.SyncConflict, local *domain.TodoChange) error {
	localData, err := json.Marshal(local)
	if err != nil {
		return err
	}

	var remoteData []byte
	if conflict.ServerVersion != nil {
		remoteData, err = json.Marshal(conflict.ServerVersion)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO sync_conflicts (user_id, entity_type, entity_id, local_version, remote_version,
		                           local_data, remote_data, resolution_strategy, status,
		                           client_timestamp, server_timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
	`

	now := time.Now()
	var remoteVersion int
	if conflict.ServerVersion != nil {
		remoteVersion = conflict.ServerVersion.Version
	}

	_, err = tx.Exec(
		context.Background(),
		query,
		userID,
		conflict.EntityType,
		conflict.EntityID,
		local.Version,
		remoteVersion,
		localData,
		remoteData,
		"manual",
		"pending",
		local.Timestamp.PhysicalTime,
		now.UnixMilli(),
		now,
	)

	return err
}

func (r *syncRepository) RecordConflict(conflict *domain.SyncConflictRecord) error {
	query := `
		INSERT INTO sync_conflicts (user_id, entity_type, entity_id, local_version, remote_version,
		                           local_data, remote_data, resolution_strategy, status,
		                           client_timestamp, server_timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
	`

	now := time.Now()
	_, err := r.db.Exec(
		context.Background(),
		query,
		conflict.UserID,
		conflict.EntityType,
		conflict.EntityID,
		conflict.LocalVersion,
		conflict.RemoteVersion,
		conflict.LocalData,
		conflict.RemoteData,
		conflict.ResolutionStrategy,
		conflict.Status,
		conflict.ClientTimestamp.PhysicalTime,
		now.UnixMilli(),
		now,
	)

	return err
}

func (r *syncRepository) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
