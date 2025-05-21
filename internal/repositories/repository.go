package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	CreateTask(task *tasks.Task) (*tasks.Task, error)

	GetAllTasks() (tasks.Tasks, error)
	GetTaskByID(id int) (*tasks.Task, error)
	GetTaskByDue(due time.Time) (*tasks.Task, error)
	GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error)

	GetTaskGoogleID(id int) (string, error)
	GetTaskIDByGoogleID(googleId string) (int, error)

	UpdateTask(task *tasks.Task) (*tasks.Task, error)

	DeleteTaskByID(id int) error
	DeleteTaskByTitle(title string) error

	ToggleCompleted(id int, completed bool) error
	MarkAsNotified(id int) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	stmt, err := r.db.Prepare("INSERT INTO tasks (google_id, title, description, completed, due, last_modified) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare create task statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		task.GoogleID,
		task.Title,
		task.Description,
		false,
		task.Due.Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create task statement for task '%s': %w", task.Title, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last insert ID for task '%s': %w", task.Title, err)
	}
	task.ID = int(id)
	return task, nil
}

func (r *taskRepository) GetAllTasks() (tasks.Tasks, error) {
	rows, err := r.db.Query("SELECT id, google_id, title, description, due, completed, notified, last_modified FROM tasks ORDER BY completed, due")
	if err != nil {
		return nil, fmt.Errorf("failed to query all tasks: %w", err)
	}
	defer rows.Close()

	var tasksList tasks.Tasks
	for rows.Next() {
		var task tasks.Task
		var dueStr string
		var lastModifiedStr string
		if err := rows.Scan(&task.ID, &task.GoogleID, &task.Title, &task.Description, &dueStr, &task.Completed, &task.Notified, &lastModifiedStr); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		if err := task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
			return nil, fmt.Errorf("failed to set due/last_modified for task ID %d: %w", task.ID, err)
		}
		tasksList = append(tasksList, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}
	return tasksList, nil
}

func (r *taskRepository) GetTaskByID(id int) (*tasks.Task, error) {
	row := r.db.QueryRow("SELECT id, google_id, title, description, due, completed, notified, last_modified FROM tasks WHERE id = ?", id)

	var task tasks.Task
	var dueStr string
	var lastModifiedStr string
	err := row.Scan(&task.ID, &task.GoogleID, &task.Title, &task.Description, &dueStr, &task.Completed, &task.Notified, &lastModifiedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task with ID %d not found: %w", id, ErrTaskNotFound)
		}
		return nil, fmt.Errorf("failed to scan task row for ID %d: %w", id, err)
	}
	if err := task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
		return nil, fmt.Errorf("failed to set due/last_modified for task ID %d: %w", id, err)
	}
	return &task, nil
}

func (r *taskRepository) GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error) {
	rows, err := r.db.Query("SELECT id, google_id, title, description, due, completed, notified, last_modified FROM tasks WHERE due >= ? AND due <= ? AND completed = 0 AND notified = 0 ORDER BY due",
		minTime.Format(time.RFC3339), maxTime.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to query pending tasks: %w", err)
	}
	defer rows.Close()

	var tasksList tasks.Tasks
	for rows.Next() {
		var task tasks.Task
		var dueStr string
		var lastModifiedStr string
		if err = rows.Scan(&task.ID, &task.GoogleID, &task.Title, &task.Description, &dueStr, &task.Completed, &task.Notified, &lastModifiedStr); err != nil {
			return nil, fmt.Errorf("failed to scan pending task row: %w", err)
		}
		if err = task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
			return nil, fmt.Errorf("failed to set due/last_modified for pending task ID %d: %w", task.ID, err)
		}
		tasksList = append(tasksList, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending task rows: %w", err)
	}
	return tasksList, nil
}

func (r *taskRepository) GetTaskByDue(due time.Time) (*tasks.Task, error) {
	row := r.db.QueryRow("SELECT id, google_id, title, description, due, completed, notified, last_modified FROM tasks WHERE due = ? LIMIT 1", due.Format(time.RFC3339))

	var task tasks.Task
	var dueStr string
	var lastModifiedStr string
	err := row.Scan(&task.ID, &task.GoogleID, &task.Title, &task.Description, &dueStr, &task.Completed, &task.Notified, &lastModifiedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task with due date %s not found: %w", due.Format(time.RFC3339), ErrTaskNotFound)
		}
		return nil, fmt.Errorf("failed to scan task row for due date %s: %w", due.Format(time.RFC3339), err)
	}
	if err = task.SetDueAndLastModified(dueStr, lastModifiedStr); err != nil {
		return nil, fmt.Errorf("failed to set due/last_modified for task with due date %s: %w", due.Format(time.RFC3339), err)
	}
	return &task, nil
}

func (r *taskRepository) GetTaskGoogleID(id int) (string, error) {
	row := r.db.QueryRow("SELECT google_id FROM tasks WHERE id = ?", id)

	var googleId sql.NullString
	err := row.Scan(&googleId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("task with ID %d not found when retrieving Google ID: %w", id, ErrTaskNotFound)
		}
		return "", fmt.Errorf("failed to retrieve Google ID for task ID %d: %w", id, err)
	}
	if !googleId.Valid {
		return "", fmt.Errorf("google ID is NULL for task ID %d", id)
	}
	return googleId.String, nil
}

func (r *taskRepository) GetTaskIDByGoogleID(googleId string) (int, error) {
	if googleId == "" {
		return 0, errors.New("google ID cannot be empty")
	}
	row := r.db.QueryRow("SELECT id FROM tasks WHERE google_id = ?", googleId)

	var id int
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("task with Google ID '%s' not found: %w", googleId, ErrTaskNotFound)
		}
		return 0, fmt.Errorf("failed to get ID by Google ID '%s': %w", googleId, err)
	}
	return id, nil
}

func (r *taskRepository) UpdateTask(task *tasks.Task) (*tasks.Task, error) {
	stmt, err := r.db.Prepare("UPDATE tasks SET google_id = ?, title = ?, description = ?, due = ?, completed = ?, notified = ?, last_modified = ? WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update task statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(task.GoogleID, task.Title, task.Description, task.Due.Format(time.RFC3339), task.Completed, task.Notified, time.Now().Format(time.RFC3339), task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update task statement for ID %d: %w", task.ID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected after updating task ID %d: %w", task.ID, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("task with ID %d not found for update: %w", task.ID, ErrTaskNotFound)
	}
	return task, nil
}

func (r *taskRepository) ToggleCompleted(id int, completed bool) error {
	stmt, err := r.db.Prepare("UPDATE tasks SET completed = ?, last_modified = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare toggle completed statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(completed, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to execute toggle completed for task ID %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for toggle completed on task ID %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with ID %d not found for toggle completed: %w", id, ErrTaskNotFound)
	}
	return nil
}

func (r *taskRepository) MarkAsNotified(id int) error {
	stmt, err := r.db.Prepare("UPDATE tasks SET notified = 1, last_modified = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare mark as notified statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to execute mark as notified for task ID %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for mark as notified on task ID %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with ID %d not found for mark as notified: %w", id, ErrTaskNotFound)
	}
	return nil
}

func (r *taskRepository) DeleteTaskByID(id int) error {
	stmt, err := r.db.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete task by ID statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to execute delete task by ID %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting task ID %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with ID %d not found for deletion: %w", id, ErrTaskNotFound)
	}
	return nil
}

func (r *taskRepository) DeleteTaskByTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty for deletion")
	}
	stmt, err := r.db.Prepare("DELETE FROM tasks WHERE title = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete task by title statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(title)
	if err != nil {
		return fmt.Errorf("failed to execute delete task by title '%s': %w", title, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting task by title '%s': %w", title, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with title '%s' not found for deletion: %w", title, ErrTaskNotFound)
	}
	return nil
}
