package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

type TaskRepository interface {
	CreateTask(task *tasks.Task) (*tasks.Task, error)
	GetAllTasks() (tasks.Tasks, error)
	GetTaskByID(id int) (*tasks.Task, error)
	GetTaskByDue(due time.Time) (tasks.Task, error)
	GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error)
	UpdateTask(task *tasks.Task) (*tasks.Task, error)
	DeleteTaskByID(id int) error
	DeleteTaskByTitle(title string) error
	ToggleCompleted(id int, completed bool) error
	MarkAsNotified(id int) error
	GetTaskGoogleID(id int) (string, error)
	GetTaskIDByGoogleID(googleId string) (int, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	row, err := r.db.Exec(
		"INSERT INTO tasks (google_id, title, description, completed, due, last_modified) VALUES (?, ?, ?, ?, ?, ?)",
		task.GoogleID,
		task.Title,
		task.Description,
		false,
		task.Due.Format(time.RFC3339),
		time.Now().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("unable to create task '%s': %w", task.Title, err)
	}

	id, err := row.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("unable to get last insert id of task %s: %w", task.Title, err)
	}
	task.ID = int(id)
	return task, nil
}

func (r *taskRepository) ToggleCompleted(id int, completed bool) error {
	_, err := r.db.Exec("UPDATE tasks SET completed = ?, last_modified = ? WHERE id = ?", completed, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("unable to toggle complete for task ID %d: %w", id, err)
	}
	return nil
}

func (r *taskRepository) GetAllTasks() (tasks.Tasks, error) {
	rows, err := r.db.Query("SELECT id, title, description, due, completed, last_modified FROM tasks ORDER BY completed, due")
	if err != nil {
		return nil, fmt.Errorf("unable to get all tasks: %w", err)
	}
	defer rows.Close()

	var tasksList tasks.Tasks
	for rows.Next() {
		var task tasks.Task
		var dueStr string
		var lastModifiedStr string
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed, &lastModifiedStr)
		if err != nil {
			return nil, err
		}
		err = task.SetDueAndLastModified(dueStr, lastModifiedStr)
		if err != nil {
			return nil, err
		}

		tasksList = append(tasksList, task)
	}
	return tasksList, nil
}

func (r *taskRepository) GetTaskGoogleID(id int) (string, error) {
	row := r.db.QueryRow("SELECT google_id FROM tasks WHERE id = ?", id)

	var googleId string
	err := row.Scan(&googleId)
	if err != nil {
		return "", fmt.Errorf("unable to get google id by id %d: %w", id, err)
	}

	return googleId, nil
}

func (r *taskRepository) GetTaskIDByGoogleID(googleId string) (int, error) {
	row := r.db.QueryRow("SELECT id FROM tasks WHERE google_id = ?", googleId)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("unable to get id by google id '%s': %w", googleId, err)
	}

	return id, nil
}

func (r *taskRepository) GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error) {
	rows, err := r.db.Query("SELECT id, title, description, due, completed, notified, last_modified FROM tasks WHERE due >= ? AND due <= ? AND notified = 0", minTime.Format(time.RFC3339), maxTime.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasksList tasks.Tasks
	for rows.Next() {
		var task tasks.Task
		var dueStr string
		var lastModifiedStr string
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed, &task.Notified, &lastModifiedStr)
		if err != nil {
			return nil, err
		}
		err = task.SetDueAndLastModified(dueStr, lastModifiedStr)
		if err != nil {
			return nil, err
		}
		tasksList = append(tasksList, task)
	}

	return tasksList, nil
}

func (r *taskRepository) GetTaskByDue(due time.Time) (tasks.Task, error) {
	row := r.db.QueryRow("SELECT id, title, description, due, completed, last_modified FROM tasks WHERE due = ?", due.Format(time.RFC3339))

	var dueStr string
	var task tasks.Task
	var lastModifiedStr string
	err := row.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed, &lastModifiedStr)
	if err != nil {
		return task, err
	}
	err = task.SetDueAndLastModified(dueStr, lastModifiedStr)
	if err != nil {
		return task, err
	}

	return task, nil
}

// TODO: update func

func (r *taskRepository) GetTaskByID(id int) (*tasks.Task, error) {
	row := r.db.QueryRow("SELECT id, title, description, due, completed, last_modified FROM tasks WHERE id = ?", id)

	var task tasks.Task
	var dueStr string
	var lastModifiedStr string
	err := row.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed, &lastModifiedStr)
	if err != nil {
		return nil, err
	}
	err = task.SetDueAndLastModified(dueStr, lastModifiedStr)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *taskRepository) UpdateTask(task *tasks.Task) (*tasks.Task, error) {
	res, err := r.db.Exec("UPDATE tasks SET title = ?, description = ?, due = ?, completed = ?, last_modified = ? WHERE id = ?",
		task.Title, task.Description, task.Due.Format(time.RFC3339), task.Completed, time.Now().Format(time.RFC3339), task.ID)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected < 1 {
		return nil, errors.New("task was not found")
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) DeleteTaskByID(id int) error {
	res, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected < 1 {
		return errors.New("task was not found")
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *taskRepository) MarkAsNotified(id int) error {
	res, err := r.db.Exec("UPDATE tasks SET notified = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected < 1 {
		return errors.New("task was not found")
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *taskRepository) DeleteTaskByTitle(title string) error {
	res, err := r.db.Exec("DELETE FROM tasks WHERE title = ?", title)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected < 1 {
		return err
	}
	return nil
}
