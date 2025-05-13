package tasks

import (
	"database/sql"
	"errors"
	"time"
)

type TaskRepository interface {
	Create(task string, description string, due time.Time) error
	GetAll() (Tasks, error)
	GetByID(id int) (*Task, error)
	GetByDue(due time.Time) (Task, error)
	Update(fields ...string) (*Task, error)
	DeleteByID(id int) error
	DeleteByTitle(title string) error
	Toogle(id int, completed bool) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task string, description string, due time.Time) error {
	_, err := r.db.Exec("INSERT INTO tasks (title, description, completed, due) VALUES (?, ?, ?, ?)", task, description, false, due.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

func (r *taskRepository) Toogle(id int, completed bool) error {
	_, err := r.db.Exec("UPDATE tasks SET completed = ? WHERE id = ?", !completed, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *taskRepository) GetAll() (Tasks, error) {
	rows, err := r.db.Query("SELECT id, title, description, due, completed FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks Tasks
	for rows.Next() {
		var task Task
		var dueStr string
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed)
		if err != nil {
			return nil, err
		}
		err = task.SetDue(dueStr)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *taskRepository) GetByDue(due time.Time) (Task, error) {
	row := r.db.QueryRow("SELECT id, title, description, due, completed FROM tasks WHERE due = ?", due.Format(time.RFC3339))

	var dueStr string
	var task Task
	err := row.Scan(&task.ID, &task.Title, &task.Description, &dueStr, &task.Completed)
	if err != nil {
		return task, err
	}
	err = task.SetDue(dueStr)
	if err != nil {
		return task, err
	}

	return task, nil
}

// TODO: update func

func (r *taskRepository) GetByID(id int) (*Task, error) {
	row := r.db.QueryRow("SELECT id, title, description, due, completed FROM tasks WHERE id = ?", id)

	task := &Task{}
	var dueStr string
	err := row.Scan(task.ID, task.Title, task.Description, dueStr, task.Completed)
	if err != nil {
		return nil, err
	}
	err = task.SetDue(dueStr)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *taskRepository) Update(fields ...string) (*Task, error) {
	return nil, nil
}

func (r *taskRepository) DeleteByID(id int) error {
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

func (r *taskRepository) DeleteByTitle(title string) error {
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
