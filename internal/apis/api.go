package apis

import "github.com/zeerodex/goot/internal/tasks"

type API interface {
	CreateTask(*tasks.Task) (*tasks.Task, error)
	GetAllLists() (tasks.TasksLists, error)
	GetTaskByID(id string) (tasks.Task, error)
	GetAllTasks() (tasks.Tasks, error)
	ToogleCompleted(id string, completed bool) error
	DeleteTaskByID(id string) error
}
