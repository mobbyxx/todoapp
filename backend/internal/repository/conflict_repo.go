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

type conflictRepository struct {
	db *pgxpool.Pool
}

func NewConflictRepository(db *pgxpool.Pool) domain.ConflictRepository {
	return &conflictRepository{db: db}
}

func (r *conflictRepository) GetByID(id uuid.UUID) (*domain.SyncConflictRecord, error) {
	query := `
		SELECT id, user_id, entity_type, entity_id, local_version, remote_version,
		       local_data, remote_data, resolution_strategy, resolved_data,
		       resolved_at, resolved_by, status, client_timestamp, server_timestamp,
		       created_at, updated_at
		FROM sync_conflicts
		WHERE id = $1
	`

	var record domain.SyncConflictRecord
	var resolvedAt *time.Time
	var resolvedBy *uuid.UUID
	var resolvedData []byte
	var clientPhysical, clientLogical int64
	var serverPhysical, serverLogical int64

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.EntityType,
		&record.EntityID,
		&record.LocalVersion,
		&record.RemoteVersion,
		&record.LocalData,
		&record.RemoteData,
		&record.ResolutionStrategy,
		&resolvedData,
		&resolvedAt,
		&resolvedBy,
		&record.Status,
		&clientPhysical,
		&clientLogical,
		&serverPhysical,
		&serverLogical,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	record.ResolvedData = resolvedData
	record.ResolvedAt = resolvedAt
	record.ResolvedBy = resolvedBy
	record.ClientTimestamp = domain.HLC{
		PhysicalTime: clientPhysical,
		LogicalTime:  clientLogical,
	}
	record.ServerTimestamp = domain.HLC{
		PhysicalTime: serverPhysical,
		LogicalTime:  serverLogical,
	}

	return &record, nil
}

func (r *conflictRepository) GetUnresolvedByUser(userID uuid.UUID) ([]*domain.SyncConflictRecord, error) {
	query := `
		SELECT id, user_id, entity_type, entity_id, local_version, remote_version,
		       local_data, remote_data, resolution_strategy, resolved_data,
		       resolved_at, resolved_by, status, client_timestamp, server_timestamp,
		       created_at, updated_at
		FROM sync_conflicts
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []*domain.SyncConflictRecord
	for rows.Next() {
		var record domain.SyncConflictRecord
		var resolvedAt *time.Time
		var resolvedBy *uuid.UUID
		var resolvedData []byte
		var clientPhysical, clientLogical int64
		var serverPhysical, serverLogical int64

		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.EntityType,
			&record.EntityID,
			&record.LocalVersion,
			&record.RemoteVersion,
			&record.LocalData,
			&record.RemoteData,
			&record.ResolutionStrategy,
			&resolvedData,
			&resolvedAt,
			&resolvedBy,
			&record.Status,
			&clientPhysical,
			&clientLogical,
			&serverPhysical,
			&serverLogical,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		record.ResolvedData = resolvedData
		record.ResolvedAt = resolvedAt
		record.ResolvedBy = resolvedBy
		record.ClientTimestamp = domain.HLC{
			PhysicalTime: clientPhysical,
			LogicalTime:  clientLogical,
		}
		record.ServerTimestamp = domain.HLC{
			PhysicalTime: serverPhysical,
			LogicalTime:  serverLogical,
		}

		conflicts = append(conflicts, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conflicts, nil
}

func (r *conflictRepository) UpdateResolution(
	conflictID uuid.UUID,
	resolution *domain.ConflictResolution,
	resolvedBy uuid.UUID,
) error {
	query := `
		UPDATE sync_conflicts
		SET resolved_data = $1,
		    resolution_strategy = $2,
		    resolved_at = $3,
		    resolved_by = $4,
		    status = 'resolved',
		    updated_at = $3
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.Exec(
		context.Background(),
		query,
		resolution.ResolvedData,
		resolution.Strategy,
		now,
		resolvedBy,
		conflictID,
	)

	return err
}

func (r *conflictRepository) RecordConflict(conflict *domain.SyncConflictRecord) error {
	query := `
		INSERT INTO sync_conflicts (
			id, user_id, entity_type, entity_id, local_version, remote_version,
			local_data, remote_data, resolution_strategy, status,
			client_timestamp, server_timestamp, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
		ON CONFLICT (id) DO UPDATE SET
			local_version = EXCLUDED.local_version,
			remote_version = EXCLUDED.remote_version,
			local_data = EXCLUDED.local_data,
			remote_data = EXCLUDED.remote_data,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	if conflict.ID == uuid.Nil {
		conflict.ID = uuid.New()
	}
	now := time.Now()
	if conflict.CreatedAt.IsZero() {
		conflict.CreatedAt = now
	}
	if conflict.UpdatedAt.IsZero() {
		conflict.UpdatedAt = now
	}

	localData, err := json.Marshal(conflict.LocalData)
	if err != nil {
		return err
	}

	remoteData, err := json.Marshal(conflict.RemoteData)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		context.Background(),
		query,
		conflict.ID,
		conflict.UserID,
		conflict.EntityType,
		conflict.EntityID,
		conflict.LocalVersion,
		conflict.RemoteVersion,
		localData,
		remoteData,
		conflict.ResolutionStrategy,
		conflict.Status,
		conflict.ClientTimestamp.PhysicalTime,
		conflict.ServerTimestamp.PhysicalTime,
		now,
	)

	return err
}
