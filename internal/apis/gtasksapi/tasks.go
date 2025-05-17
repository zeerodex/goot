package gtasksapi

import (
	"fmt"
	"time"

	gtasks "google.golang.org/api/tasks/v1"

	"github.com/zeerodex/goot/internal/tasks"
)

type GTasksApi struct {
	srv    *gtasks.Service
	ListId string
}

func NewGTasksApi(listId string) *GTasksApi {
	return &GTasksApi{srv: GetService(), ListId: listId}
}

func (api *GTasksApi) InsertTaskIntoList(task *gtasks.Task) error {
	_, err := api.srv.Tasks.Insert(api.ListId, task).Do()
	if err != nil {
		return fmt.Errorf("error inserting task into %s: %v", api.ListId, err)
	}
	return nil
}

func (api *GTasksApi) GetTasksLists() (*gtasks.TaskLists, error) {
	lists, err := api.srv.Tasklists.List().MaxResults(10).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching task lists: %v", err)
	}
	return lists, nil
}

func (api *GTasksApi) GetTasks() (*gtasks.Tasks, error) {
	tasks, err := api.srv.Tasks.List(api.ListId).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching tasks from %s: %v", api.ListId, err)
	}
	return tasks, nil
}

func (api *GTasksApi) DeleteTaskById(taskId string) error {
	err := api.srv.Tasks.Delete(api.ListId, taskId).Do()
	if err != nil {
		return fmt.Errorf("error deleting task %s from default list: %v", taskId, err)
	}
	return nil
}

func ConvertGTask(g *gtasks.Task) (*tasks.Task, error) {
	var t tasks.Task
	var err error

	t.GoogleID = g.Id
	t.Title = g.Title
	t.Description = g.Notes
	if g.Due != "" {
		t.Due, err = time.Parse(time.RFC3339, g.Due)
		if err != nil {
			return nil, fmt.Errorf("error parsing due: %w", err)
		}
	}
	if g.Status == "completed" {
		t.Completed = true
	} else {
		t.Completed = false
	}

	return &t, nil
}
