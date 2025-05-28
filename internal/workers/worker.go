package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

type Worker struct {
	ID       int
	jobQueue <-chan APIJob
	resultCh chan<- APIJobResult

	gApi apis.API
	apis []apis.API
	repo repositories.TaskRepository
}

func NewWorker(id int, jobChan <-chan APIJob, resChan chan<- APIJobResult, apis []apis.API, repo repositories.TaskRepository) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobChan,
		resultCh: resChan,

		gApi: apis[0],
		apis: apis,
		repo: repo,
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-w.jobQueue:
			if !ok {
				return
			}
			result := w.processAPIJob(job)

			select {
			case w.resultCh <- result:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) processAPIJob(job APIJob) APIJobResult {
	var err error
	switch job.Operation {
	case ToggleCompletedOp:
		err = w.processToggleCompletedTaskOp(job.TaskID, job.Completed)
	case UpdateTaskOp:
		err = w.processUpdateTaskOp(job.Task)
	case DeleteTaskOp:
		err = w.processDeleteTaskOp(job.TaskID)
	case CreateTaskOp:
		err = w.processCreateTaskOp(job.Task)
	case SyncTasksOp:
		err = w.processSyncTasksOp()
	}

	res := APIJobResult{
		JobID:     job.ID,
		Operation: job.Operation,
		Success:   err == nil,
		Err:       err,
	}

	return res
}

func (w *Worker) processDeleteTaskOp(id int) error {
	for _, api := range w.apis {
		googleId, err := w.repo.GetTaskGoogleID(id)
		if err != nil {
			return err
		}
		err = api.DeleteTaskByID(googleId)
		if err != nil {
			return err
		}

	}
	return nil
}

func (w *Worker) processCreateTaskOp(task *tasks.Task) error {
	for _, api := range w.apis {
		gtask, err := api.CreateTask(task)
		if err != nil {
			return err
		}

		err = w.repo.UpdateGoogleID(task.ID, gtask.GoogleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) processUpdateTaskOp(task *tasks.Task) error {
	for _, api := range w.apis {
		if task.GoogleID == "" {
			return fmt.Errorf("task ID %d has no google id, run sync to add task in gtasks", task.ID)
		} else {
			_, err := api.PatchTask(task)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Worker) processToggleCompletedTaskOp(id int, completed bool) error {
	for _, api := range w.apis {
		googleId, err := w.repo.GetTaskGoogleID(id)
		if err != nil {
			return err
		}
		err = api.ToggleCompleted(googleId, completed)
		if err != nil {
			return err
		}

	}
	return nil
}

func (w *Worker) processSyncTasksOp() error {
	err := w.SyncGTasks()
	return err
}
