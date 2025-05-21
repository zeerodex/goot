package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zeerodex/goot/internal/apis"
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

	SyncGTasks() error

	GetGApi() apis.API
}

type taskService struct {
	repo repositories.TaskRepository

	gApi  apis.API
	gSync bool
}

func NewTaskService(repo repositories.TaskRepository, gApi apis.API, gSync bool) TaskService {
	return &taskService{repo: repo, gApi: gApi, gSync: gSync}
}

func (s *taskService) GetGApi() apis.API {
	return s.gApi
}

func (s *taskService) SyncGTasks() error {
	gtasks, err := s.gApi.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all tasks from Google API: %w", err)
	}

	for _, gtask := range gtasks {
		gtask.ID, err = s.repo.GetTaskIDByGoogleID(gtask.GoogleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_, err := s.repo.CreateTask(&gtask)
				if err != nil {
					return fmt.Errorf("failed to create local task for Google ID '%s': %w", gtask.GoogleID, err)
				}
				log.Printf("Successfully createad and synced local task ID %d for Google ID %s", gtask.ID, gtask.GoogleID)
			} else {
				return fmt.Errorf("faield to get local task ID for GoogleID '%s': %w", gtask.GoogleID, err)
			}
		}
		task, err := s.repo.GetTaskByID(gtask.ID)
		if err != nil {
			return fmt.Errorf("failed to get local task by ID %d (for Google ID '%s'): %w", gtask.ID, gtask.GoogleID, err)
		}
		timeDiff := gtask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			switch timeDiff {
			case 1:
				_, err := s.gApi.PatchTask(task)
				if err != nil {
					return fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", gtask.GoogleID, gtask.ID, err)
				}
			case -1:
				_, err := s.repo.UpdateTask(&gtask)
				if err != nil {
					return fmt.Errorf("failed to update local task (ID %d) with newer googla task (Google ID '%s'): %w", gtask.ID, gtask.GoogleID, err)
				}
			}
		}
	}

	return nil
}

func (s *taskService) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	if s.gSync {
		_, err := s.gApi.CreateTask(task)
		if err != nil {
			return nil, fmt.Errorf("failed to create task in Google API: %w", err)
		}
	}

	task, err := s.repo.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task in repository: %w", err)
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
	if s.gSync {
		id, err := s.repo.GetTaskGoogleID(id)
		if err != nil {
			return err
		}
		err = s.gApi.ToggleCompleted(id, completed)
		if err != nil {
			return err
		}
	}
	return s.repo.ToggleCompleted(id, completed)
}

func (s *taskService) DeleteTaskByID(id int) error {
	if s.gSync {
		googleId, err := s.repo.GetTaskGoogleID(id)
		if err != nil {
			return err
		}
		err = s.gApi.DeleteTaskByID(googleId)
		if err != nil {
			return err
		}
	}
	err := s.repo.DeleteTaskByID(id)
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
