package services

import (
	"fmt"
	"log"
	"time"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/apis/gtasksapi"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

type TaskService interface {
	CreateTask(task *tasks.Task) (*tasks.Task, error)
	GetTaskByID(id int) (*tasks.Task, error)
	GetAllTasks() (tasks.Tasks, error)
	GetAllPendingTasks(minTime, maxTime time.Time) (tasks.Tasks, error)
	ToggleCompleted(id int, completed bool) error
	MarkAsNotified(id int) error
	DeleteTaskByID(id int) error
	UpdateTask(task *tasks.Task) (*tasks.Task, error)

	Sync() error

	GetGApi() apis.API
}

type taskService struct {
	repo repositories.TaskRepository

	gApi apis.API

	cfg *config.Config
}

func NewTaskService(repo repositories.TaskRepository, cfg *config.Config) (TaskService, error) {
	var gApi apis.API
	for api, enabled := range cfg.APIs {
		if api == "google" && enabled {
			srv, err := gtasksapi.GetService()
			if err != nil {
				return nil, fmt.Errorf("failed to enable Google API: %v", err)
			}
			gApi = gtasksapi.NewGTasksApi(srv, cfg.Google.ListId)
		}
	}
	return &taskService{repo: repo, gApi: gApi, cfg: cfg}, nil
}

func (s *taskService) GetGApi() apis.API {
	return s.gApi
}

func (s *taskService) Sync() error {
	var err error
	if s.cfg.Google.Sync {
		err = s.SyncGTasks()
	}
	if err != nil {
		return fmt.Errorf("failed to sync apis: %w", err)
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

	if s.cfg.Google.Sync {
		go func(task *tasks.Task) {
			if task.GoogleID != "" {
				_, err := s.gApi.PatchTask(task)
				if err != nil {
					log.Println(err)
				}
			}
		}(task)
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

	if s.cfg.Google.Sync {
		go func(task *tasks.Task) {
			gtask, err := s.gApi.CreateTask(task)
			if err != nil {
				log.Println(err)
			}

			err = s.repo.UpdateGoogleID(task.ID, gtask.GoogleID)
			if err != nil {
				log.Println(err)
			}
		}(task)
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

func (s *taskService) ToggleCompleted(id int, completed bool) error {
	if s.cfg.Google.Sync {
		go func(id int, completed bool) {
			googleId, err := s.repo.GetTaskGoogleID(id)
			if err != nil {
				log.Println(err)
			}
			err = s.gApi.ToggleCompleted(googleId, completed)
			if err != nil {
				log.Println(err)
			}
		}(id, completed)
	}

	return s.repo.ToggleCompleted(id, completed)
}

func (s *taskService) DeleteTaskByID(id int) error {
	if s.cfg.Google.Sync {
		go func(id int) {
			googleId, err := s.repo.GetTaskGoogleID(id)
			if err != nil {
				log.Println(err)
			}
			err = s.gApi.DeleteTaskByID(googleId)
			if err != nil {
				log.Println(err)
			}
		}(id)
	}

	err := s.repo.SoftDeleteTaskByID(id)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) GetTaskGoogleID(id int) (string, error) {
	return s.repo.GetTaskGoogleID(id)
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
