package gtasksapi

import (
	"fmt"
	"time"

	gtasks "google.golang.org/api/tasks/v1"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/tasks"
)

type GTasksApi struct {
	srv *gtasks.Service

	ListId string
}

func NewGTasksApi(srv *gtasks.Service, listId string) apis.API {
	return &GTasksApi{srv: srv, ListId: listId}
}

func (api *GTasksApi) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	gtask, err := api.srv.Tasks.Insert(api.ListId, task.GTask()).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create task in list '%s': %w", api.ListId, err)
	}
	task.GoogleID = gtask.Id
	return task, nil
}

func (api *GTasksApi) GetTaskByID(id string) (*tasks.Task, error) {
	gtask, err := api.srv.Tasks.Get(api.ListId, id).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve task '%s' from list '%s': %w", id, api.ListId, err)
	}
	return ConvertGTask(gtask), nil
}

func (api *GTasksApi) GetAllLists() (tasks.TasksLists, error) {
	glists, err := api.srv.Tasklists.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all task lists: %w", err)
	}
	lists := make(tasks.TasksLists, len(glists.Items))
	for i, glist := range glists.Items {
		lists[i] = ConverGTasksList(glist)
	}
	return lists, nil
}

func (api *GTasksApi) GetAllTasks() (tasks.Tasks, error) {
	gtasks, err := api.srv.Tasks.List(api.ListId).ShowCompleted(true).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all tasks from list '%s': %w", api.ListId, err)
	}

	tasksList := make(tasks.Tasks, len(gtasks.Items))
	for i, task := range gtasks.Items {
		t := ConvertGTask(task)
		tasksList[i] = *t
	}
	return tasksList, nil
}

func (api *GTasksApi) GetAllTasksWithDeleted() (tasks.Tasks, error) {
	gtasks, err := api.srv.Tasks.List(api.ListId).ShowDeleted(true).ShowCompleted(true).ShowHidden(true).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all deleted tasks from list '%s': %w", api.ListId, err)
	}

	tasksList := make(tasks.Tasks, len(gtasks.Items))
	for i, task := range gtasks.Items {
		t := ConvertGTask(task)
		tasksList[i] = *t
	}
	return tasksList, nil
}

func (api *GTasksApi) DeleteTaskByID(id string) error {
	err := api.srv.Tasks.Delete(api.ListId, id).Do()
	if err != nil {
		return fmt.Errorf("failed to delete task '%s' from list '%s': %w", id, api.ListId, err)
	}
	return nil
}

func (api *GTasksApi) PatchTask(task *tasks.Task) (*tasks.Task, error) {
	g, err := api.srv.Tasks.Patch(api.ListId, task.GoogleID, task.GTask()).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to patch task '%s' from list '%s': %w", task.GoogleID, api.ListId, err)
	}
	return ConvertGTask(g), nil
}

func (api *GTasksApi) ToggleCompleted(id string, completed bool) error {
	gtask := &gtasks.Task{}
	if completed {
		gtask.Status = "completed"
	} else {
		gtask.Status = "needsAction"
	}

	_, err := api.srv.Tasks.Patch(api.ListId, id, gtask).Do()
	if err != nil {
		return fmt.Errorf("failed to delete task '%s' from list '%s': %v", id, api.ListId, err)
	}
	return nil
}

func ConvertGTask(g *gtasks.Task) *tasks.Task {
	t := &tasks.Task{
		GoogleID:    g.Id,
		Title:       g.Title,
		Description: g.Notes,
		Completed:   g.Status == "completed",
		Deleted:     g.Deleted,
	}
	if g.Due != "" {
		t.Due, _ = time.Parse(time.RFC3339, g.Due)
	}
	t.LastModified, _ = time.Parse(time.RFC3339, g.Updated)
	if g.Completed != nil {
		completedTime, _ := time.Parse(time.RFC3339, *g.Completed)
		timeDiff := t.LastModified.Compare(completedTime)
		if timeDiff == -1 {
			t.LastModified = completedTime
		}
	}
	return t
}

func ConverGTasksList(g *gtasks.TaskList) (t tasks.TasksList) {
	t.Title = g.Title
	t.GoogleID = g.Id
	return t
}
