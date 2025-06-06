package workers

import (
	"context"
	"sync"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/tasks"
)

type Worker struct {
	ID       int
	jobQueue <-chan APIJob
	resultCh chan<- APIJobResult

	apis map[string]apis.API
	repo repositories.TaskRepository
}

func NewWorker(id int, jobChan <-chan APIJob, resChan chan<- APIJobResult, apis map[string]apis.API, repo repositories.TaskRepository) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobChan,
		resultCh: resChan,

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

// TODO: implement retry logic
func (w *Worker) processAPIJob(job APIJob) APIJobResult {
	var err error
	switch job.Operation {
	case SetTaskCompletedOp:
		err = w.processSetTaskCompletedOp(job.TaskID, job.Completed)
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
		TaskID:    job.TaskID,
		Success:   err == nil,
		Err:       err,
	}

	return res
}

func (w *Worker) processDeleteTaskOp(id int) error {
	for apiName, api := range w.apis {
		apiId, err := w.repo.GetTaskAPIID(id, apiName)
		if err != nil {
			return err
		}
		err = api.DeleteTaskByID(apiId)
		if err != nil {
			return err
		}

	}
	return nil
}

func (w *Worker) processCreateTaskOp(task *tasks.Task) error {
	for apiName, api := range w.apis {
		apiTask, err := api.CreateTask(task)
		if err != nil {
			return err
		}

		var apiId string
		switch apiName {
		case "gtasks":
			apiId = apiTask.GoogleID
		case "todoist":
			apiId = apiTask.TodoistID
		}

		err = w.repo.UpdateTaskAPIID(task.ID, apiId, apiName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) processUpdateTaskOp(task *tasks.Task) error {
	for _, api := range w.apis {
		_, err := api.PatchTask(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) processSetTaskCompletedOp(id int, completed bool) error {
	for apiName, api := range w.apis {
		apiId, err := w.repo.GetTaskAPIID(id, apiName)
		if err != nil {
			return err
		}

		err = api.SetTaskCompleted(apiId, completed)
		if err != nil {
			return err
		}

	}
	return nil
}

func (w *Worker) processSyncTasksOp() error {
	err := w.SyncAPITasks()
	return err
}
