package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepo struct {
	db *sql.DB
}

func (r *PostgresRepo) FindByID(ctx context.Context, id int64) (*Task, error) {
	t, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Create(ctx context.Context, title string) (Task, error) {
	const query = `
		INSERT INTO	tasks (title)
		VALUES ($1)
		RETURNING id, title, done, created_at, updated_at
	`
	var task Task

	err := r.db.QueryRowContext(ctx, query, title).Scan(
		&task.ID,
		&task.Title,
		&task.Done,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]Task, error) {
	const query = `
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Done,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepo) Get(ctx context.Context, id int64) (Task, error) {
	const query = `
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task Task

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Done,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepo) Update(
	ctx context.Context,
	id int64,
	title string,
	done bool,
) (Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			done = $2,
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, title, done, created_at, updated_at
	`

	var task Task

	err := r.db.QueryRowContext(ctx, query, title, done, id).Scan(
		&task.ID,
		&task.Title,
		&task.Done,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM tasks
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted row count: %w", err)
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepo) ListDone(
	ctx context.Context,
	done bool,
) ([]Task, error) {
	const query = `
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		WHERE done = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, done)
	if err != nil {
		return nil, fmt.Errorf("list tasks by done: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Done,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate filtered tasks: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepo) CreateMany(
	ctx context.Context,
	titles []string,
) error {
	if len(titles) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO tasks (title)
		VALUES ($1)
	`

	for _, title := range titles {
		if _, err := tx.ExecContext(ctx, query, title); err != nil {
			return fmt.Errorf("insert task in transaction: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
