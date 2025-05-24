package services

import (
	"fmt"
	"sync"

	"github.com/zeerodex/goot/internal/tasks"
)

func (s *taskService) processMissingGTasks(gtasks, ltasks tasks.Tasks) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(ltasks))

	semaphore := make(chan struct{}, 5)

	for _, task := range ltasks {
		if _, found := gtasks.FindTaskByGoogleID(task.GoogleID); found || task.Deleted {
			continue
		}

		wg.Add(1)
		go func(t tasks.Task) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			_, err := s.gApi.CreateTask(&task)
			if err != nil {
				errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", task.GoogleID, err)
				return
			}
			gtasks = append(gtasks, task)

			err = s.repo.UpdateGoogleID(task.ID, task.GoogleID)
			if err != nil {
				errChan <- fmt.Errorf("failed to update Google ID '%s' of task ID %d: %w", task.GoogleID, task.ID, err)
				return
			}
		}(task)

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

	fmt.Println(gtasks)
	for _, gtask := range gtasks {
		task, found := ltasks.FindTaskByGoogleID(gtask.GoogleID)
		if !found {
			if gtask.Deleted {
				continue
			}
			wg.Add(1)
			go func(gt tasks.Task) {
				defer wg.Done()

				var err error
				task, err = s.repo.CreateTask(&gtask)
				if err != nil {
					errChan <- fmt.Errorf("failed to create local task for Google ID '%s': %w", gtask.GoogleID, err)
					return
				}
			}(gtask)
			continue
		}

		gtask.ID = task.ID

		if gtask.Deleted {
			err := s.repo.DeleteTaskByID(task.ID)
			if err != nil {
				return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", gtask.ID, err)
			}
			return nil
		}
		if task.Deleted {
			err := s.gApi.DeleteTaskByID(gtask.GoogleID)
			if err != nil {
				return fmt.Errorf("failed to delete marked as deleted task ID %d: %w", task.ID, err)
			}
			return nil
		}

		timeDiff := gtask.LastModified.Compare(task.LastModified)
		if timeDiff != 0 {
			switch timeDiff {
			case -1:
				wg.Add(1)
				go func(t tasks.Task) {
					defer wg.Done()
					_, err := s.gApi.PatchTask(&t)
					if err != nil {
						errChan <- fmt.Errorf("failed to patch google task (Google ID '%s') with newer local task (ID %d): %w", task.GoogleID, task.ID, err)
						return
					}
				}(*task)
			case 1:
				wg.Add(1)
				go func(gt tasks.Task) {
					defer wg.Done()
					_, err := s.repo.UpdateTask(&gt)
					if err != nil {
						errChan <- fmt.Errorf("failed to update local task (ID %d) with newer Google task (Google ID '%s'): %w", gtask.ID, gtask.GoogleID, err)
						return
					}
				}(gtask)
			}
		}
	}
	go func() {
		wg.Done()
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
