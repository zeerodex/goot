package services

import (
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/apis/gtasksapi"
	"github.com/zeerodex/goot/internal/apis/todoist"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/workers"
)

type TaskService interface {
	CreateTask(task *tasks.Task) (*tasks.Task, error)
	GetTaskByID(id int) (*tasks.Task, error)
	GetAllTasks() (tasks.Tasks, error)
	GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error)
	SetTaskCompleted(id int, completed bool) error
	MarkAsNotified(id int) error
	DeleteTaskByID(id int) error
	UpdateTask(task *tasks.Task) (*tasks.Task, error)

	Sync() error

	WP() *workers.APIWorkerPool
}

type taskService struct {
	repo repositories.TaskRepository

	cfg *config.Config

	gApi apis.API
	wp   *workers.APIWorkerPool
}

func NewTaskService(repo repositories.TaskRepository, cfg *config.Config) (TaskService, error) {
	apisMap := make(map[string]apis.API)
	for api, enabled := range cfg.APIs {
		if api == "google" && enabled {
			googleAPI, err := gtasksapi.NewGTasksApi(cfg.Google.ListId)
			if err != nil {
				return nil, fmt.Errorf("failed to enable Google API: %v", err)
			}
			apisMap["gtasks"] = googleAPI
		}
		if api == "todoist" && enabled {
			todoistAPI, err := todoist.NewTodoistAPI()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize Todoist API: %w", err)
			}
			apisMap["todoist"] = todoistAPI
		}
	}

	wp := workers.NewAPIWorkerPool(3, 5, apisMap, repo)
	wp.Start()

	return &taskService{repo: repo, cfg: cfg, wp: wp}, nil
}

func (s *taskService) WP() *workers.APIWorkerPool {
	return s.wp
}

func (s *taskService) Sync() error {
	err := s.wp.Submit(workers.APIJob{
		Operation: workers.SyncTasksOp,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) UpdateTask(task *tasks.Task) (*tasks.Task, error) {
	if err := s.ValidateTask(task); err != nil {
		return nil, fmt.Errorf("unable to validate task: %w", err)
	}

	task, err := s.repo.UpdateTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to update local task ID %d: %w", task.ID, err)
	}

	err = s.wp.Submit(workers.APIJob{
		Operation: workers.UpdateTaskOp,
		Task:      task,
		TaskID:    task.ID,
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	if err := s.ValidateTask(task); err != nil {
		return nil, fmt.Errorf("unable to validate task: %w", err)
	}
	task, err := s.repo.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task in repository: %w", err)
	}

	err = s.wp.Submit(workers.APIJob{
		Operation: workers.CreateTaskOp,
		Task:      task,
		TaskID:    task.ID,
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTaskByID(id int) (*tasks.Task, error) {
	return s.repo.GetTaskByID(id)
}

func (s *taskService) GetAllTasks() (tasks.Tasks, error) {
	return s.repo.GetAllTasks()
}

func (s *taskService) GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error) {
	return s.repo.GetAllPendingTasks(minTime, maxTime)
}

func (s *taskService) SetTaskCompleted(id int, completed bool) error {
	if err := s.repo.SetTaskCompleted(id, completed); err != nil {
		return err
	}

	err := s.wp.Submit(workers.APIJob{
		Operation: workers.SetTaskCompletedOp,
		TaskID:    id,
		Completed: completed,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *taskService) DeleteTaskByID(id int) error {
	err := s.repo.SoftDeleteTaskByID(id)
	if err != nil {
		return err
	}

	err = s.wp.Submit(workers.APIJob{
		Operation: workers.DeleteTaskOp,
		TaskID:    id,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *taskService) GetTaskGoogleID(id int) (string, error) {
	return s.repo.GetTaskAPIID(id, "gtasks")
}

func (s *taskService) MarkAsNotified(id int) error {
	return s.repo.MarkAsNotified(id)
}

func (s *taskService) ValidateTask(task *tasks.Task) error {
	if len(task.Title) > s.cfg.MaxLength.Title {
		return fmt.Errorf("allowed length of task title - %d", s.cfg.MaxLength.Title)
	}
	if len(task.Description) > s.cfg.MaxLength.Description {
		return fmt.Errorf("allowed length of task description - %d", s.cfg.MaxLength.Description)
	}
	return nil
}
