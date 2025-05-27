package workers

import (
	"context"
	"fmt"
	"log"
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
	repo repositories.TaskRepository
}

func NewWorker(id int, jobChan <-chan APIJob, resChan chan<- APIJobResult, gApi apis.API, repo repositories.TaskRepository) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobChan,
		resultCh: resChan,

		gApi: gApi,
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

			return
		case <-ctx.Done():
			log.Printf("API Worker %d stopping due to context cancellation", w.ID)
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
		JobID:        job.ID,
		Operation:    job.Operation,
		TaskID:       job.Task.ID,
		TaskGoogleID: job.TaskGoogleID,
		Success:      err == nil,
		Err:          err,
	}

	return res
}

func (w *Worker) processDeleteTaskOp(id int) error {
	googleId, err := w.repo.GetTaskGoogleID(id)
	if err != nil {
		return err
	}
	err = w.gApi.DeleteTaskByID(googleId)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) processCreateTaskOp(task *tasks.Task) error {
	gtask, err := w.gApi.CreateTask(task)
	if err != nil {
		return err
	}

	err = w.repo.UpdateGoogleID(task.ID, gtask.GoogleID)
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) processUpdateTaskOp(task *tasks.Task) error {
	if task.GoogleID == "" {
		return fmt.Errorf("task ID %d has no google id, run sync to add task in gtasks", task.ID)
	} else {
		_, err := w.gApi.PatchTask(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) processToggleCompletedTaskOp(id int, completed bool) error {
	googleId, err := w.repo.GetTaskGoogleID(id)
	if err != nil {
		return err
	}
	err = w.gApi.ToggleCompleted(googleId, completed)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) processSyncTasksOp() error {
	err := w.SyncGTasks()
	return err
}
