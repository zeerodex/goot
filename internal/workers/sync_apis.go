package workers

import (
	"fmt"
	"sync"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

const semaphoreSize = 3

func processMissingAPITasks(atasks, ltasks tasks.Tasks, api apis.API, repo repositories.TaskRepository) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(ltasks))
	semaphore := make(chan struct{}, semaphoreSize)

	for _, task := range ltasks {
		if _, found := atasks.FindTaskByGoogleID(task.GoogleID); found || task.Deleted {
			continue
		}

		wg.Add(1)
		go func(t *tasks.Task) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			_, err := api.CreateTask(t)
			if err != nil {
				errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", t.GoogleID, err)
				return
			}

			ltasks = append(ltasks, *t)

			err = repo.UpdateGoogleID(t.ID, t.GoogleID)
			if err != nil {
				errChan <- fmt.Errorf("failed to update Google ID '%s' of task ID %d: %w", t.GoogleID, t.ID, err)
				return
			}
		}(&task)

	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func processMissingLocalTasks(atasks, ltasks tasks.Tasks, api apis.API, repo repositories.TaskRepository) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(atasks))
	semaphore := make(chan struct{}, semaphoreSize)

	for _, atask := range atasks {
		task, found := ltasks.FindTaskByGoogleID(atask.GoogleID)
		if !found {
			if atask.Deleted {
				continue
			}
			wg.Add(1)
			go func(at *tasks.Task) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				var err error
				task, err = repo.CreateTask(at)
				if err != nil {
					errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", atask.GoogleID, err)
					return
				}
			}(&atask)
			continue
		}

		if task.Deleted || atask.Deleted {
			wg.Add(1)
			go func(tId int, atId string, tDel, atDel bool) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if atDel {
					err := repo.DeleteTaskByID(tId)
					if err != nil {
						errChan <- fmt.Errorf("failed to delete marked as deleted local task ID %d: %w", tId, err)
						return
					}
				}
				if tDel {
					err := api.DeleteTaskByID(atId)
					if err != nil {
						errChan <- fmt.Errorf("failed to delete marked as deleted google task Google ID '%s': %w", atId, err)
						return
					}
				}
			}(task.ID, atask.GoogleID, task.Deleted, atask.Deleted)

		}

		timeDiff := atask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			atask.ID = task.ID
			switch timeDiff {
			case -1:
				wg.Add(1)
				go func(t *tasks.Task) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					_, err := api.PatchTask(t)
					if err != nil {
						errChan <- fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
						return
					}
				}(task)
			case 1:
				wg.Add(1)
				go func(at *tasks.Task) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					_, err := repo.UpdateTask(at)
					if err != nil {
						errChan <- fmt.Errorf("failed to update local task (ID %d) with newer Google task (Google ID '%s'): %w", task.ID, atask.GoogleID, err)
						return
					}
				}(&atask)
			}
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) SyncAPITasks() error {
	for _, api := range w.apis {
		var wg sync.WaitGroup
		errChan := make(chan error, 2)

		var tasks, deletedTasks, atasks tasks.Tasks

		var err error
		wg.Add(1)
		go func() {
			defer wg.Done()

			tasks, err = w.repo.GetAllTasks()
			if err != nil {
				errChan <- fmt.Errorf("failed to get all local tasks: %w", err)
				return
			}
			deletedTasks, err = w.repo.GetAllDeletedTasks()
			if err != nil {
				errChan <- fmt.Errorf("failed to get all deleted local tasks: %w", err)
				return
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			atasks, err = api.GetAllTasksWithDeleted()
			if err != nil {
				errChan <- fmt.Errorf("failed to get all google tasks: %w", err)
				return
			}
		}()

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return err
			}
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
