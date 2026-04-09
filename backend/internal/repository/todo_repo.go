package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type todoRepository struct {
	db *pgxpool.Pool
}

func NewTodoRepository(db *pgxpool.Pool) domain.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *domain.Todo) error {
	query := `
		INSERT INTO todos (title, description, status, priority, created_by, assigned_to, due_date, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now
	todo.Version = 1

	if todo.Status == "" {
		todo.Status = domain.TodoStatusPending
	}
	if todo.Priority == "" {
		todo.Priority = domain.TodoPriorityMedium
	}

	return r.db.QueryRow(
		context.Background(),
		query,
		todo.Title,
		todo.Description,
		todo.Status,
		todo.Priority,
		todo.CreatedBy,
		todo.AssignedTo,
		todo.DueDate,
		todo.Version,
		todo.CreatedAt,
		todo.UpdatedAt,
	).Scan(&todo.ID)
}

func (r *todoRepository) GetByID(id uuid.UUID) (*domain.Todo, error) {
	query := `
		SELECT id, title, description, status, priority, created_by, assigned_to, due_date, version, created_at, updated_at, deleted_at
		FROM todos
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanTodo(r.db.QueryRow(context.Background(), query, id))
}

func (r *todoRepository) GetByUserID(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, error) {
	query := `
		SELECT id, title, description, status, priority, created_by, assigned_to, due_date, version, created_at, updated_at, deleted_at
		FROM todos
		WHERE deleted_at IS NULL AND (created_by = $1 OR assigned_to = $1)
	`
	args := []interface{}{userID}
	argCount := 1

	if filters.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
	}

	if filters.Priority != nil {
		argCount++
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filters.Priority)
	}

	if filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filters.Search+"%")
	}

	query += " ORDER BY created_at DESC"

	if filters.PageSize > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.PageSize)

		if filters.Page > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, (filters.Page-1)*filters.PageSize)
		}
	}

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTodos(rows)
}

func (r *todoRepository) Update(todo *domain.Todo) error {
	query := `
		UPDATE todos
		SET title = $1, description = $2, status = $3, priority = $4, assigned_to = $5, 
		    due_date = $6, version = version + 1, updated_at = $7
		WHERE id = $8 AND version = $9 AND deleted_at IS NULL
		RETURNING version
	`

	todo.UpdatedAt = time.Now()

	var newVersion int
	err := r.db.QueryRow(
		context.Background(),
		query,
		todo.Title,
		todo.Description,
		todo.Status,
		todo.Priority,
		todo.AssignedTo,
		todo.DueDate,
		todo.UpdatedAt,
		todo.ID,
		todo.Version,
	).Scan(&newVersion)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existingTodo, getErr := r.GetByID(todo.ID)
			if getErr != nil {
				if errors.Is(getErr, domain.ErrTodoNotFound) {
					return domain.ErrTodoNotFound
				}
				return getErr
			}
			if existingTodo != nil && existingTodo.Version != todo.Version {
				return domain.ErrVersionMismatch
			}
			return domain.ErrTodoNotFound
		}
		return err
	}

	todo.Version = newVersion
	return nil
}

func (r *todoRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE todos
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}

	return nil
}

func (r *todoRepository) List(filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argCount := 0

	if filters.UserID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND (created_by = $%d OR assigned_to = $%d)", argCount, argCount)
		args = append(args, *filters.UserID)
	}

	if filters.Status != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
	}

	if filters.Priority != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filters.Priority)
	}

	if filters.AssignedTo != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND assigned_to = $%d", argCount)
		args = append(args, *filters.AssignedTo)
	}

	if filters.Search != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filters.Search+"%")
	}

	countQuery := "SELECT COUNT(*) FROM todos " + whereClause
	var totalCount int
	err := r.db.QueryRow(context.Background(), countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, title, description, status, priority, created_by, assigned_to, due_date, version, created_at, updated_at, deleted_at
		FROM todos
	` + whereClause + " ORDER BY created_at DESC"

	if filters.PageSize > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.PageSize)

		if filters.Page > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, (filters.Page-1)*filters.PageSize)
		}
	}

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	todos, err := r.scanTodos(rows)
	if err != nil {
		return nil, 0, err
	}

	return todos, totalCount, nil
}

func (r *todoRepository) scanTodo(row pgx.Row) (*domain.Todo, error) {
	var todo domain.Todo
	var assignedTo *uuid.UUID
	var dueDate, deletedAt *time.Time

	err := row.Scan(
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
			return nil, domain.ErrTodoNotFound
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

func (r *todoRepository) scanTodos(rows pgx.Rows) ([]*domain.Todo, error) {
	var todos []*domain.Todo

	for rows.Next() {
		var todo domain.Todo
		var assignedTo *uuid.UUID
		var dueDate, deletedAt *time.Time

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

		todos = append(todos, &todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func sanitizeSearchTerm(term string) string {
	term = strings.ReplaceAll(term, "%", "\\%")
	term = strings.ReplaceAll(term, "_", "\\_")
	return term
}
