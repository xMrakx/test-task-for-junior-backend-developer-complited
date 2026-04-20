package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func (r *Repository) GetByType(ctx context.Context, taskType taskdomain.Type) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, task_type, task_interval, schedule_at, created_at, updated_at
		FROM tasks
		WHERE task_type = $1
		ORDER BY id DESC
	`

	row, err := r.pool.Query(ctx, query, taskType)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	tasks := make([]taskdomain.Task, 0)
	for row.Next() {
		task, err := scanTask(row)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, task_type, task_interval, schedule_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, task_type, task_interval, schedule_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.Type, task.Interval, task.ScheduleAt, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, task_type, task_interval, schedule_at, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			task_type = $4,
			task_interval = $5,
			schedule_at = $6,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, task_type, task_interval, schedule_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.Type, task.Interval, task.ScheduleAt, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, task_type, task_interval, schedule_at, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.Type,
		&task.Interval,
		&task.ScheduleAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
