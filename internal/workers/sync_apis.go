package workers

import (
	"errors"
	"fmt"

	"github.com/zeerodex/goot/internal/models"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

func (w *Worker) Sync() error {
	ltasks, err := w.repo.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all local tasks: %w", err)
	}
	deletedLTasks, err := w.repo.GetAllDeletedTasks()
	if err != nil {
		return fmt.Errorf("failed to get all deleted tasks: %w", err)
	}
	ltasks = append(ltasks, deletedLTasks...)

	apisTasks := make(map[string]tasks.Tasks)
	snapshots := make(models.Snapshots)
	for apiName, api := range w.apis {
		apisTasks[apiName], err = api.GetAllTasks()
		if err != nil {
			return fmt.Errorf("failed to get all tasks from %s api: %w", apiName, err)
		}
		snapshots[apiName], err = w.snapRepo.GetLastSnapshot(apiName)
		if err != nil {
			if !errors.Is(err, repositories.ErrSnapshotNotFound) {
				return fmt.Errorf("failed to get snapshot for %s api: %w", apiName, err)
			}
		}
	}

	if apisTasksEqual(apisTasks) {
		apiTasks := mergeAPIsTasks(apisTasks)
		if slicesEqual(ltasks, apiTasks) {
			return nil
		} else {
			err = w.syncLTasks(ltasks, apiTasks)
			if err != nil {
				return fmt.Errorf("failed to replace local tasks with api tasks: %w", err)
			}
		}
	}

	return nil
}

func (w *Worker) syncLTasks(ltasks tasks.Tasks, apiTasks tasks.Tasks) error {
	var err error
	for _, apiTask := range apiTasks {
		task, found := ltasks.FindTaskByAPIID(apiTask.APIIDs[apiTask.Source], apiTask.Source)
		if found {
			if task.Deleted {
				for apiName, api := range w.apis {
					err = api.DeleteTaskByID(apiTask.APIIDs[apiName])
					if err != nil {
						return fmt.Errorf("local task was deleted: failed to delete task from %s api: %w", apiName, err)
					}
				}
				err = w.repo.DeleteTaskByID(task.ID)
				if err != nil {
					return fmt.Errorf("failed to delete task: %w", err)
				}
				continue
			}
			if !task.Equal(apiTask) {
				switch apiTask.LastModified.Compare(task.LastModified) {
				case 1:
					apiTask.ID = task.ID
					if _, err := w.repo.UpdateTask(&apiTask); err != nil {
						return fmt.Errorf("failed to update task api id: %w", err)
					}
				case -1:
					for _, api := range w.apis {
						if _, err := api.UpdateTask(task); err != nil {
							return fmt.Errorf("failed to update task: %w", err)
						}
					}
				}
			}
		} else {
			_, err = w.repo.CreateTask(&apiTask)
			if err != nil {
				return fmt.Errorf("local task was not found: failed to create local task: %w", err)
			}
			continue
		}
	}
	return nil
}

func apisTasksEqual(apisTasks map[string]tasks.Tasks) bool {
	if len(apisTasks) == 0 {
		return true
	}

	var reference tasks.Tasks
	for _, slice := range apisTasks {
		reference = slice
		break
	}

	for _, slice := range apisTasks {
		if !slicesEqual(reference, slice) {
			return false
		}
	}
	return true
}

func slicesEqual(slice1, slice2 tasks.Tasks) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if !slice1[i].Equal(slice2[i]) || slice1[i].Deleted != slice2[i].Deleted {
			return false
		}
	}
	return true
}

// All tasks slices must be equal
func mergeAPIsTasks(apisTasks map[string]tasks.Tasks) tasks.Tasks {
	var templateTasks tasks.Tasks
	for _, slice := range apisTasks {
		templateTasks = slice
		break
	}

	mergedTasks := make(tasks.Tasks, len(templateTasks))

	for i, task := range templateTasks {
		for apiName, apiTasks := range apisTasks {
			task.APIIDs[apiName] = apiTasks[i].APIIDs[apiName]
		}
		mergedTasks[i] = task
	}
	return mergedTasks
}
