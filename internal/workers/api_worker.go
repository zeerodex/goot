package workers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

type APIOperation string

const (
	CreateTaskOp      APIOperation = "create_task"
	UpdateTaskOp      APIOperation = "update_task"
	DeleteTaskOp      APIOperation = "delete_task"
	ToggleCompletedOp APIOperation = "toggle_completed_task"
	SyncTasksOp       APIOperation = "sync_tasks"
)

type APIJob struct {
	ID           int
	Operation    APIOperation
	Task         *tasks.Task
	TaskID       int
	TaskGoogleID string
	Retry        int
	Completed    bool
}

type APIJobResult struct {
	JobID        int
	Operation    APIOperation
	TaskID       int
	TaskGoogleID string
	Success      bool
	Err          error
}

type APIWorkerPool struct {
	numWorkers int
	workers    []*Worker
	jobs       chan APIJob
	results    chan APIJobResult
	quit       chan bool
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewAPIWorkerPool(numWorkers int, queueSize int, gApi apis.API, repo repositories.TaskRepository) *APIWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &APIWorkerPool{
		numWorkers: numWorkers,
		workers:    make([]*Worker, numWorkers),
		jobs:       make(chan APIJob, queueSize),
		results:    make(chan APIJobResult, queueSize),
		quit:       make(chan bool),
		ctx:        ctx,
		cancel:     cancel,
	}

	for i := range wp.numWorkers {
		wp.workers[i] = NewWorker(i, wp.jobs, wp.results, wp.quit, gApi, repo)
	}
	return wp
}

func (wp *APIWorkerPool) Submit(job APIJob) error {
	select {
	case wp.jobs <- job:
		return nil
	case <-wp.ctx.Done():
		return errors.New("failed to submit job due to context cancellation")
	default:
		return errors.New("failed to submit job")

	}
}

func (wp *APIWorkerPool) Stop() {
	wp.cancel()
	close(wp.jobs)
	for range wp.workers {
		select {
		case wp.quit <- true:
		default:
		}
	}
	wp.wg.Wait()
	close(wp.results)
	close(wp.quit)
}

func (wp *APIWorkerPool) Result() <-chan APIJobResult {
	return wp.results
}

func NewWorker(id int, jobChan <-chan APIJob, resChan chan<- APIJobResult, quitChan <-chan bool, gApi apis.API, repo repositories.TaskRepository) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobChan,
		resultCh: resChan,
		quit:     quitChan,

		gApi: gApi,
		repo: repo,
	}
}

type Worker struct {
	ID       int
	jobQueue <-chan APIJob
	resultCh chan<- APIJobResult
	quit     <-chan bool

	gApi apis.API
	repo repositories.TaskRepository
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

		case <-w.quit:
			log.Printf("API Worker %d stopping", w.ID)
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
