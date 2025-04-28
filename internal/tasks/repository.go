package tasks

import "database/sql"

type TaskRepository interface {
	Create(task string, description string) error
	GetAll() (Tasks, error)
	GetByID(id int) (*Task, error)
	Update(fields ...string) (*Task, error)
	DeleteByID(id int) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task string, description string) error {
	_, err := r.db.Exec("INSERT INTO tasks (task, description) VALUES (?, ?)", task, description)
	if err != nil {
		return err
	}
	return nil
}

func (r *taskRepository) GetAll() (Tasks, error) {
	rows, err := r.db.Query("SELECT id, task, description, status FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks Tasks
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Task, &task.Description, &task.Status)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

// TODO: update func

func (r *taskRepository) GetByID(id int) (*Task, error) {
	row := r.db.QueryRow("SELECT id, task, description FROM tasks WHERE id = ?", id)

	task := &Task{}
	err := row.Scan(task.ID, task.Task, task.Description)
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
		return err
	}
	return nil
}
