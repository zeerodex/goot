package services

import (
	"fmt"
	"sync"

	"github.com/zeerodex/goot/internal/tasks"
)

const semaphoreSize = 3

func (s *taskService) processMissingGTasks(gtasks, ltasks tasks.Tasks) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(ltasks))
	semaphore := make(chan struct{}, semaphoreSize)

	for _, task := range ltasks {
		if _, found := gtasks.FindTaskByGoogleID(task.GoogleID); found || task.Deleted {
			continue
		}

		wg.Add(1)
		go func(t *tasks.Task) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			_, err := s.gApi.CreateTask(t)
			if err != nil {
				errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", t.GoogleID, err)
				return
			}

			ltasks = append(ltasks, *t)

			err = s.repo.UpdateGoogleID(t.ID, t.GoogleID)
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

func (s *taskService) processMissingLocalTasks(gtasks, ltasks tasks.Tasks) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(gtasks))
	semaphore := make(chan struct{}, semaphoreSize)

	for _, gtask := range gtasks {
		task, found := ltasks.FindTaskByGoogleID(gtask.GoogleID)
		if !found {
			if gtask.Deleted {
				continue
			}
			wg.Add(1)
			go func(gt *tasks.Task) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				var err error
				task, err = s.repo.CreateTask(gt)
				if err != nil {
					errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", gtask.GoogleID, err)
					return
				}
			}(&gtask)
			continue
		}

		if task.Deleted || gtask.Deleted {
			wg.Add(1)
			go func(tId int, gtId string, tDel, gtDel bool) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if gtDel {
					err := s.repo.DeleteTaskByID(tId)
					if err != nil {
						errChan <- fmt.Errorf("failed to delete marked as deleted local task ID %d: %w", tId, err)
						return
					}
				}
				if tDel {
					err := s.gApi.DeleteTaskByID(gtId)
					if err != nil {
						errChan <- fmt.Errorf("failed to delete marked as deleted google task Google ID '%s': %w", gtId, err)
						return
					}
				}
			}(task.ID, gtask.GoogleID, task.Deleted, gtask.Deleted)

		}

		timeDiff := gtask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			gtask.ID = task.ID
			switch timeDiff {
			case -1:
				wg.Add(1)
				go func(t *tasks.Task) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					_, err := s.gApi.PatchTask(t)
					if err != nil {
						errChan <- fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
						return
					}
				}(task)
			case 1:
				wg.Add(1)
				go func(gt *tasks.Task) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					_, err := s.repo.UpdateTask(gt)
					if err != nil {
						errChan <- fmt.Errorf("failed to update local task (ID %d) with newer Google task (Google ID '%s'): %w", task.ID, gtask.GoogleID, err)
						return
					}
				}(&gtask)
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

func (s *taskService) SyncGTasks() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	var tasks, deletedTasks, gtasks tasks.Tasks

	var err error
	wg.Add(1)
	go func() {
		defer wg.Done()

		tasks, err = s.repo.GetAllTasks()
		if err != nil {
			errChan <- fmt.Errorf("failed to get all local tasks: %w", err)
			return
		}
		deletedTasks, err = s.repo.GetAllDeletedTasks()
		if err != nil {
			errChan <- fmt.Errorf("failed to get all deleted local tasks: %w", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		gtasks, err = s.gApi.GetAllTasksWithDeleted()
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

	if err = s.processMissingGTasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failed to process missing google tasks: %w", err)
	}

	if err = s.processMissingLocalTasks(gtasks, tasks); err != nil {
		return fmt.Errorf("failed to process missing local tasks: %w", err)
	}

	return nil
}
