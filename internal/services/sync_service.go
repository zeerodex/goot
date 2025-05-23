package services

import (
	"fmt"

	"github.com/zeerodex/goot/internal/tasks"
)

func (s *taskService) processMissingGTasks(gtasks, tasks tasks.Tasks) error {
	for _, task := range tasks {
		_, found := gtasks.FindTaskByGoogleID(task.GoogleID)
		if !found && !task.Deleted {
			_, err := s.gApi.CreateTask(&task)
			if err != nil {
				return fmt.Errorf("failed to create local task for Google ID '%s': %w", task.GoogleID, err)
			}
			gtasks = append(gtasks, task)

			err = s.repo.UpdateGoogleID(task.ID, task.GoogleID)
			if err != nil {
				return fmt.Errorf("failed to update Google ID '%s' of task ID %d: %w", task.GoogleID, task.ID, err)
			}
		}
	}
	return nil
}

func (s *taskService) processMissingLocalTasks(gtasks, tasks tasks.Tasks) error {
	for _, gtask := range gtasks {
		task, found := tasks.FindTaskByGoogleID(gtask.GoogleID)
		if !found {
			if gtask.Deleted {
				continue
			}
			var err error
			task, err = s.repo.CreateTask(&gtask)
			if err != nil {
				return fmt.Errorf("failed to create local task for Google ID '%s': %w", gtask.GoogleID, err)
			}
			tasks = append(tasks, *task)
			continue
		}

		gtask.ID = task.ID

		if gtask.Deleted {
			err := s.repo.DeleteTaskByID(gtask.ID)
			if err != nil {
				return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", gtask.ID, err)
			}
			return nil
		}
		if task.Deleted {
			err := s.gApi.DeleteTaskByID(task.GoogleID)
			if err != nil {
				return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", task.ID, err)
			}
			return nil
		}

		timeDiff := gtask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			switch timeDiff {
			case -1:
				_, err := s.gApi.PatchTask(task)
				if err != nil {
					return fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
				}
			case 1:
				_, err := s.repo.UpdateTask(&gtask)
				if err != nil {
					return fmt.Errorf("failed to update local task (ID %d) with newer Google task (Google ID '%s'): %w", gtask.ID, gtask.GoogleID, err)
				}
			}
		}
	}
	return nil
}

func (s *taskService) SyncGTasks() error {
	tasks, err := s.repo.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all local tasks: %w", err)
	}
	deletedTasks, err := s.repo.GetAllDeletedTasks()
	if err != nil {
		return fmt.Errorf("failed to get all deleted local tasks: %w", err)
	}
	tasks = append(tasks, deletedTasks...)

	gtasks, err := s.gApi.GetAllTasksWithDeleted()
	if err != nil {
		return fmt.Errorf("failed to get all google tasks: %w", err)
	}

	if err = s.processMissingGTasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failed to process missing google tasks: %w", err)
	}

	if err = s.processMissingLocalTasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failed to process missing local tasks: %w", err)
	}

	return nil
}
