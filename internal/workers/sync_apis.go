package workers

import (
	"fmt"
	"time"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

func processMissingAPITasks(atasks, ltasks tasks.Tasks, api apis.API, repo repositories.TaskRepository) error {
	for _, task := range ltasks {
		if _, found := atasks.FindTaskByGoogleID(task.GoogleID); found || task.Deleted {
			continue
		}

		_, err := api.CreateTask(&task)
		if err != nil {
			return fmt.Errorf("failed to create local task for Google ID '%s': %w", task.GoogleID, err)
		}

		ltasks = append(ltasks, task)

		err = repo.UpdateTaskAPIID(task.ID, task.GoogleID, "gtasks")
		if err != nil {
			return fmt.Errorf("failed to update Google ID '%s' of task ID %d: %w", task.GoogleID, task.ID, err)
		}
	}

	return nil
}

func processMissingLocalTasks(atasks, ltasks tasks.Tasks, api apis.API, repo repositories.TaskRepository) error {
	for _, atask := range atasks {
		task, found := ltasks.FindTaskByGoogleID(atask.GoogleID)
		var err error
		if !found {
			if atask.Deleted {
				continue
			}

			_, err = repo.CreateTask(&atask)
			if err != nil {
				return fmt.Errorf("failed to create local task for Google ID '%s': %w", atask.GoogleID, err)
			}
			continue
		}

		if task.Deleted || atask.Deleted {
			if atask.Deleted {
				err = repo.DeleteTaskByID(task.ID)
				if err != nil {
					return fmt.Errorf("failed to delete marked as deleted local task ID %d: %w", task.ID, err)
				}
			}
			if task.Deleted {
				err = api.DeleteTaskByID(atask.GoogleID)
				if err != nil {
					return fmt.Errorf("failed to delete marked as deleted google task Google ID '%s': %w", atask.GoogleID, err)
				}
			}
		}

		timeDiff := atask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			atask.ID = task.ID
			switch timeDiff {
			case -1:
				_, err = api.PatchTask(task)
				if err != nil {
					return fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
				}
			case 1:
				if atask.Due.Truncate(24 * time.Hour).Equal(task.Due.Truncate(24 * time.Hour)) {
					atask.Due = task.Due
				}
				_, err = repo.UpdateTask(&atask)
				if err != nil {
					return fmt.Errorf("failed to update local task (ID %d) with newer Google task (Google ID '%s'): %w", task.ID, atask.GoogleID, err)
				}
			}
		}
	}

	return nil
}

func (w *Worker) SyncAPITasks() error {
	for _, api := range w.apis {
		var tasks, deletedTasks, atasks tasks.Tasks
		var err error

		tasks, err = w.repo.GetAllTasks()
		if err != nil {
			return fmt.Errorf("failed to get all local tasks: %w", err)
		}
		deletedTasks, err = w.repo.GetAllDeletedTasks()
		if err != nil {
			return fmt.Errorf("failed to get all deleted local tasks: %w", err)
		}

		atasks, err = api.GetAllTasksWithDeleted()
		if err != nil {
			return fmt.Errorf("failed to get all google tasks: %w", err)
		}

		tasks = append(tasks, deletedTasks...)

		if err = processMissingAPITasks(atasks, tasks, api, w.repo); err != nil {
			return fmt.Errorf("failed to process missing google tasks: %w", err)
		}

		if err = processMissingLocalTasks(atasks, tasks, api, w.repo); err != nil {
			return fmt.Errorf("failed to process missing local tasks: %w", err)
		}

	}
	return nil
}
