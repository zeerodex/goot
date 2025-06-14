package workers

import (
	"fmt"

	"github.com/zeerodex/goot/internal/tasks"
)

func (w *Worker) Sync() error {
	ltasks, err := w.repo.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all local tasks: %w", err)
	}

	apisTasks := make(map[string]tasks.Tasks)
	// snapshots := make(map[string]*models.Snapshot)
	for apiName, api := range w.apis {
		apisTasks[apiName], err = api.GetAllTasks()
		if err != nil {
			return fmt.Errorf("failed to get all tasks from %s api: %w", apiName, err)
		}
		// snapshots[apiName], err = w.snapRepo.GetLastSnapshot(apiName)
		// if err != nil {
		// 	return fmt.Errorf("failed to get snapshot for %s api: %w", apiName, err)
		// }
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

// HACK:
// func (w *Worker) replaceLTasksWithAPITasks(apiTasks tasks.Tasks) error {
// 	err := w.repo.DeleteAllTasks()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete all tasks: %w", err)
// 	}
// 	for _, task := range apiTasks {
// 		_, err = w.repo.CreateTask(&task)
// 		if err != nil {
// 			return fmt.Errorf("failed to create task: %w", err)
// 		}
// 	}
// 	return nil
// }

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
		if !slice1[i].Equal(slice2[i]) {
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

func (w *Worker) syncLTasks(ltasks tasks.Tasks, apiTasks tasks.Tasks) error {
	var err error
	for _, apiTask := range apiTasks {
		task, found := ltasks.FindTaskByAPIID(apiTask.APIIDs[apiTask.Source], apiTask.Source)
		if !found {
			task, err = w.repo.CreateTask(&apiTask)
			if err != nil {
				return fmt.Errorf("local task was not found: failed to create local task: %w", err)
			}
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
	}
	return nil
}
