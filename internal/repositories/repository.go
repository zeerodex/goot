package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

var ErrTaskNotFound = errors.New("task not found")

type scanner interface {
	Scan(dest ...any) error
}

type TaskRepository interface {
	CreateTask(task *tasks.Task) (*tasks.Task, error)

	GetAllTasks() (tasks.Tasks, error)
	GetTaskByID(id int) (*tasks.Task, error)
	GetTaskByDue(due time.Time) (*tasks.Task, error)
	GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error)
	GetAllDeletedTasks() (tasks.Tasks, error)

	GetTaskIDByGoogleID(googleId string) (int, error)
	GetTaskAPIID(id int, apiName string) (string, error)

	UpdateTask(task *tasks.Task) (*tasks.Task, error)
	UpdateTaskAPIID(id int, apiId string, apiName string) error

	DeleteTaskByID(id int) error
	SoftDeleteTaskByID(id int) error
	DeleteTaskByTitle(title string) error
	DeleteAllTasks() error

	SetTaskCompleted(id int, completed bool) error
	MarkAsNotified(id int) error

	DB() *sql.DB
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) DB() *sql.DB {
	return r.db
}

func (r *taskRepository) scanTask(s scanner) (*tasks.Task, error) {
	var task tasks.Task
	var dueStr, lastModifiedStr string
	err := s.Scan(
		&task.ID, &task.GoogleID, &task.TodoistID, &task.Title,
		&task.Description, &dueStr, &task.Completed, &task.Notified,
		&lastModifiedStr, &task.Deleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to scan task row: %w", err)
	}
	if err := task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
		return nil, fmt.Errorf("failed to parse time strings: %w", err)
	}
	return &task, nil
}

func (r *taskRepository) findTasks(query string, args ...any) (tasks.Tasks, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasksList tasks.Tasks
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasksList = append(tasksList, *task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}
	return tasksList, nil
}

func (r *taskRepository) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	query := "INSERT INTO tasks (google_id, todoist_id, title, description, due, completed, last_modified) VALUES (?, ?, ?, ?, ?, ?, ?)"
	now := time.Now().UTC().Format(time.RFC3339)
	due := task.Due.Format(time.RFC3339)

	res, err := r.db.Exec(query, task.GoogleID, task.TodoistID, task.Title, task.Description, due, task.Completed, now)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create task statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	task.ID = int(id)
	return task, nil
}

const selectAllFields = "SELECT id, google_id, todoist_id, title, description, due, completed, notified, last_modified, deleted FROM tasks"

func (r *taskRepository) GetAllTasks() (tasks.Tasks, error) {
	query := fmt.Sprintf("%s WHERE deleted = 0 ORDER BY completed, due", selectAllFields)
	return r.findTasks(query)
}

func (r *taskRepository) GetAllDeletedTasks() (tasks.Tasks, error) {
	query := fmt.Sprintf("%s WHERE deleted = 1 ORDER BY completed, due", selectAllFields)
	return r.findTasks(query)
}

func (r *taskRepository) GetTaskByID(id int) (*tasks.Task, error) {
	query := fmt.Sprintf("%s WHERE id = ?", selectAllFields)
	row := r.db.QueryRow(query, id)
	return r.scanTask(row)
}

func (r *taskRepository) GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error) {
	query := fmt.Sprintf("%s WHERE due >= ? AND due <= ? AND completed = 0 AND notified = 0 ORDER BY due", selectAllFields)
	return r.findTasks(query, minTime.Format(time.RFC3339), maxTime.Format(time.RFC3339))
}

func (r *taskRepository) GetTaskByDue(due time.Time) (*tasks.Task, error) {
	query := fmt.Sprintf("%s WHERE due = ? LIMIT 1", selectAllFields)
	row := r.db.QueryRow(query, due.Format(time.RFC3339))
	return r.scanTask(row)
}

func (r *taskRepository) GetTaskIDByGoogleID(googleId string) (int, error) {
	if googleId == "" {
		return 0, errors.New("google ID cannot be empty")
	}
	var id int
	err := r.db.QueryRow("SELECT id FROM tasks WHERE google_id = ?", googleId).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("task with Google ID '%s' not found: %w", googleId, ErrTaskNotFound)
		}
		return 0, fmt.Errorf("failed to get ID by Google ID '%s': %w", googleId, err)
	}
	return id, nil
}

func (r *taskRepository) GetTaskAPIID(id int, apiName string) (string, error) {
	var fieldName string
	switch apiName {
	case "todoist":
		fieldName = "todoist_id"
	case "gtasks":
		fieldName = "google_id"
	default:
		return "", fmt.Errorf("unsupported api name: %s", apiName)
	}

	query := fmt.Sprintf("SELECT %s FROM tasks WHERE id = ?", fieldName)
	var apiId sql.NullString
	if err := r.db.QueryRow(query, id).Scan(&apiId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrTaskNotFound
		}
		return "", fmt.Errorf("failed to scan api id row: %w", err)
	}
	if !apiId.Valid {
		return "", fmt.Errorf("%s is NULL for task id %d", fieldName, id)
	}
	return apiId.String, nil
}

// execStatement is a helper for UPDATE, DELETE, and other statements that don't return rows.
func (r *taskRepository) execStatement(query string, args ...any) error {
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *taskRepository) UpdateTask(task *tasks.Task) (*tasks.Task, error) {
	query := "UPDATE tasks SET google_id = ?, todoist_id = ?, title = ?, description = ?, due = ?, completed = ?, notified = ?, last_modified = ? WHERE id = ?"
	now := time.Now().UTC().Format(time.RFC3339)
	due := task.Due.Format(time.RFC3339)

	err := r.execStatement(query, task.GoogleID, task.TodoistID, task.Title, task.Description, due, task.Completed, task.Notified, now, task.ID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) UpdateTaskAPIID(id int, apiId string, apiName string) error {
	var fieldName string
	switch apiName {
	case "todoist":
		fieldName = "todoist_id"
	case "gtasks":
		fieldName = "google_id"
	default:
		return fmt.Errorf("unsupported api name: %s", apiName)
	}
	query := fmt.Sprintf("UPDATE tasks SET %s = ?, last_modified = ? WHERE id = ?", fieldName)
	return r.execStatement(query, apiId, time.Now().UTC().Format(time.RFC3339), id)
}

func (r *taskRepository) SetTaskCompleted(id int, completed bool) error {
	query := "UPDATE tasks SET completed = ?, last_modified = ? WHERE id = ?"
	return r.execStatement(query, completed, time.Now().UTC().Format(time.RFC3339), id)
}

func (r *taskRepository) MarkAsNotified(id int) error {
	query := "UPDATE tasks SET notified = 1, last_modified = ? WHERE id = ?"
	return r.execStatement(query, time.Now().UTC().Format(time.RFC3339), id)
}

func (r *taskRepository) DeleteTaskByID(id int) error {
	return r.execStatement("DELETE FROM tasks WHERE id = ?", id)
}

func (r *taskRepository) DeleteAllTasks() error {
	return r.execStatement("DELETE FROM tasks")
}

func (r *taskRepository) SoftDeleteTaskByID(id int) error {
	query := "UPDATE tasks SET deleted = 1, last_modified = ? WHERE id = ?"
	return r.execStatement(query, time.Now().UTC().Format(time.RFC3339), id)
}

func (r *taskRepository) DeleteTaskByTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty for deletion")
	}
	return r.execStatement("DELETE FROM tasks WHERE title = ?", title)
}
