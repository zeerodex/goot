package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

type APISnapshotsRepository interface {
	CreateTask(task *tasks.APITask) (*tasks.APITask, error)
	GetAllTasks() (tasks.APITasks, error)
	GetTaskByID(id string) (*tasks.APITask, error)
	UpdateTask(task *tasks.APITask) (*tasks.APITask, error)
	DeleteTaskByID(id string) error
}

type apiSnapshotsRepository struct {
	apiName   string
	tableName string
	db        *sql.DB
}

func NewAPISnapshotsRepository(apiName string, db *sql.DB) APISnapshotsRepository {
	if db == nil {
		panic("database connection cannot be nil")
	}
	if apiName == "" {
		panic("API name cannot be empty")
	}

	return &apiSnapshotsRepository{
		apiName:   apiName,
		tableName: apiName + "_snapshots",
		db:        db,
	}
}

func (r *apiSnapshotsRepository) CreateTask(task *tasks.APITask) (*tasks.APITask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (api_id, title, description, due, completed, last_modified) 
		VALUES (?, ?, ?, ?, ?, ?)`, r.tableName)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare create task statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	_, err = stmt.Exec(
		task.APIID,
		task.Title,
		task.Description,
		task.Due.Format(time.RFC3339),
		task.Completed,
		now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task '%s': %w", task.Title, err)
	}

	return task, nil
}

func (r *apiSnapshotsRepository) GetAllTasks() (tasks.APITasks, error) {
	query := fmt.Sprintf(`
		SELECT id, api_id, title, description, due, completed, last_modified 
		FROM %s 
		ORDER BY completed ASC, due ASC`, r.tableName)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tasks: %w", err)
	}
	defer rows.Close()

	var tasksList tasks.APITasks
	for rows.Next() {
		task, err := r.scanAPITask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasksList = append(tasksList, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasksList, nil
}

func (r *apiSnapshotsRepository) GetTaskByID(id string) (*tasks.APITask, error) {
	if id == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	query := fmt.Sprintf(`
		SELECT id, api_id, title, description, due, completed, last_modified 
		FROM %s 
		WHERE id = ?`, r.tableName)

	row := r.db.QueryRow(query, id)
	task, err := r.scanAPITask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task by ID '%s': %w", id, err)
	}

	return task, nil
}

func (r *apiSnapshotsRepository) UpdateTask(task *tasks.APITask) (*tasks.APITask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}
	if task.APIID == "" {
		return nil, errors.New("task API ID cannot be empty")
	}

	query := fmt.Sprintf(`
		UPDATE %s 
		SET title = ?, description = ?, due = ?, completed = ?, last_modified = ? 
		WHERE api_id = ?`, r.tableName)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update task statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	result, err := stmt.Exec(
		task.Title,
		task.Description,
		task.Due.Format(time.RFC3339),
		task.Completed,
		now.Format(time.RFC3339),
		task.APIID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task '%s': %w", task.Title, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (r *apiSnapshotsRepository) DeleteTaskByID(id string) error {
	if id == "" {
		return errors.New("task ID cannot be empty")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.tableName)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare delete task statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete task with ID '%s': %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *apiSnapshotsRepository) scanAPITask(rows scanner) (*tasks.APITask, error) {
	var task tasks.APITask
	var dueStr, lastModifiedStr string

	err := rows.Scan(
		&task.APIID,
		&task.Title,
		&task.Description,
		&dueStr,
		&task.Completed,
		&lastModifiedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to scan task row: %w", err)
	}
	if err := task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
		return nil, fmt.Errorf("failed to parse timestamps: %w", err)
	}

	return &task, nil
}
