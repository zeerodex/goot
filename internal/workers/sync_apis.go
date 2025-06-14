package workers

//
// import (
// 	"fmt"
//
// 	"github.com/zeerodex/goot/internal/models"
// 	"github.com/zeerodex/goot/internal/tasks"
// )
//
// func (w *Worker) Sync() error {
// 	ltasks, err := w.repo.GetAllTasks()
// 	if err != nil {
// 		return fmt.Errorf("failed to get all local tasks: %w", err)
// 	}
//
// 	apisTasks := make(map[string]tasks.Tasks)
// 	var snapshots map[string]*models.Snapshot
// 	for apiName, api := range w.apis {
// 		apisTasks[apiName], err = api.GetAllTasks()
// 		if err != nil {
// 			return fmt.Errorf("failed to get all tasks from %s api: %w", apiName, err)
// 		}
// 		snapshots[apiName], err = w.snapRepo.GetLastSnapshot(apiName)
// 		if err != nil {
// 			return fmt.Errorf("failed to get snapshot for %s api: %w", apiName, err)
// 		}
// 	}
//
// 	if apisTasksEqual(apisTasks) {
// 		apiTasks := mergeAPIsTasks(apisTasks)
// 		if slicesEqual(ltasks, apiTasks) {
// 			return nil
// 		} else {
// 			err = w.replaceLTasksWithAPITasks(apiTasks)
// 			if err != nil {
// 				fmt.Println(err.Error())
// 				return fmt.Errorf("failed to replace local tasks with api tasks: %w", err)
// 			}
// 		}
// 	}
//
// 	return nil
// }
//
// // HACK:
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
//
// func apisTasksEqual(apisTasks map[string]tasks.Tasks) bool {
// 	if len(apisTasks) == 0 {
// 		return true
// 	}
//
// 	var reference tasks.Tasks
// 	for _, slice := range apisTasks {
// 		reference = slice
// 		break
// 	}
//
// 	for _, slice := range apisTasks {
// 		if !slicesEqual(reference, slice) {
// 			return false
// 		}
// 	}
// 	return true
// }
//
// func slicesEqual(slice1, slice2 tasks.Tasks) bool {
// 	if len(slice1) != len(slice2) {
// 		return false
// 	}
//
// 	for i := range slice1 {
// 		if !slice1[i].Equal(slice2[i]) {
// 			return false
// 		}
// 	}
// 	return true
// }
//
// // All tasks slices must be equal
// func mergeAPIsTasks(apisTasks map[string]tasks.Tasks) tasks.Tasks {
// 	var templateTasks tasks.Tasks
// 	for _, slice := range apisTasks {
// 		templateTasks = slice
// 		break
// 	}
//
// 	mergedTasks := make(tasks.Tasks, len(templateTasks))
//
// 	for i, task := range templateTasks {
// 		for apiName, apiTasks := range apisTasks {
// 			task.SetAPIID(apiName, apiTasks[i].GetAPIID(apiName))
// 		}
// 		mergedTasks[i] = task
// 	}
// 	return mergedTasks
// }
