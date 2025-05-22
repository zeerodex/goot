package services

import (
	"fmt"

	"github.com/zeerodex/goot/internal/tasks"
)

func (s *taskService) processMissingGtasks(gtasks, tasks tasks.Tasks) error {
	var missingIDs []int
	for _, task := range tasks {
		if _, found := gtasks.FindTaskIDByGoogleID(task.GoogleID); !found {
			missingIDs = append(missingIDs, task.ID)
		}
	}

	for _, id := range missingIDs {
		task, found := tasks.FindID(id)
		if !found {
			return fmt.Errorf("unable to find task in task list by ID %d", id)
		}

		if task.Deleted {
			continue
		}

		gtask, err := s.gApi.CreateTask(task)
		if err != nil {
			return fmt.Errorf("failed to create google task with Google ID '%s': %w", task.GoogleID, err)
		}

		gtasks = append(gtasks, *gtask)

		err = s.repo.UpdateGoogleID(task.ID, task.GoogleID)
		if err != nil {
			return fmt.Errorf("failed to update Google ID '%s' of task ID %d: %w", task.GoogleID, task.ID, err)
		}
	}
	return nil
}

func (s *taskService) processMissingLocalTasks(gtasks, tasks tasks.Tasks) error {
	for _, gtask := range gtasks {
		var found bool
		gtask.ID, found = tasks.FindTaskIDByGoogleID(gtask.GoogleID)
		if !found {
			task, err := s.repo.CreateTask(&gtask)
			if err != nil {
				return fmt.Errorf("failed to create local task for Google ID '%s': %w", gtask.GoogleID, err)
			}
			tasks = append(tasks, *task)
		}

		task, found := tasks.FindID(gtask.ID)
		if !found {
			return fmt.Errorf("failed to find task ID %d inside tasks", gtask.ID)
		}

		timeDiff := gtask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			switch timeDiff {
			case 1:
				if gtask.Deleted {
					err := s.gApi.DeleteTaskByID(task.GoogleID)
					if err != nil {
						return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", task.ID, err)
					}
				}
				_, err := s.gApi.PatchTask(task)
				if err != nil {
					return fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
				}
			case -1:
				if gtask.Deleted {
					err := s.repo.DeleteTaskByID(gtask.ID)
					if err != nil {
						return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", gtask.ID, err)
					}
				}
				_, err := s.repo.UpdateTask(&gtask)
				if err != nil {
					return fmt.Errorf("failed to update local task (ID %d) with newer googla task (Google ID '%s'): %w", gtask.ID, gtask.GoogleID, err)
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
	gtasks, err := s.gApi.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all google tasks: %w", err)
	}

	if err = s.processMissingGtasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failed to process missing google tasks: %w", err)
	}

	if err = s.processMissingLocalTasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failet ot process missing local tasks: %w", err)
	}

	return nil
}
