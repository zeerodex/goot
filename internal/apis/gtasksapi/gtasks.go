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

func NewGTasksApi(listId string) apis.API {
	return &GTasksApi{srv: GetService(), ListId: listId}
}

func (api *GTasksApi) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	gtask, err := api.srv.Tasks.Insert(api.ListId, task.GTask()).Do()
	if err != nil {
		return nil, fmt.Errorf("error inserting task into %s: %v", api.ListId, err)
	}
	task.GoogleID = gtask.Id
	return task, nil
}

func (api *GTasksApi) GetTaskByID(id string) (tasks.Task, error) {
	gtask, err := api.srv.Tasks.Get(api.ListId, id).Do()
	if err != nil {
		return tasks.Task{}, fmt.Errorf("error fetching task from %s: %v", api.ListId, err)
	}
	return ConvertGTask(gtask), nil
}

func (api *GTasksApi) GetAllLists() (tasks.TasksLists, error) {
	glists, err := api.srv.Tasklists.List().Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching task lists: %v", err)
	}
	lists := make(tasks.TasksLists, len(glists.Items))
	for i, glist := range glists.Items {
		lists[i] = ConverGTasksList(glist)
	}
	return lists, nil
}

func (api *GTasksApi) GetAllTasks() (tasks.Tasks, error) {
	gtasks, err := api.srv.Tasks.List(api.ListId).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching tasks from %s: %v", api.ListId, err)
	}

	tasksList := make(tasks.Tasks, len(gtasks.Items))
	for i, task := range gtasks.Items {
		tasksList[i] = ConvertGTask(task)
	}
	return tasksList, nil
}

func (api *GTasksApi) DeleteTaskByID(taskId string) error {
	err := api.srv.Tasks.Delete(api.ListId, taskId).Do()
	if err != nil {
		return fmt.Errorf("error deleting task %s from gtasks: %v", taskId, err)
	}
	return nil
}

func (api *GTasksApi) ToogleCompleted(id string, completed bool) error {
	gtask := &gtasks.Task{}
	if completed {
		gtask.Status = "completed"
	} else {
		gtask.Status = "needsAction"
	}

	_, err := api.srv.Tasks.Patch(api.ListId, id, gtask).Do()
	if err != nil {
		return fmt.Errorf("error deleting task %s from default list: %v", id, err)
	}
	return nil
}

func ConvertGTask(g *gtasks.Task) (t tasks.Task) {
	t.GoogleID = g.Id
	t.Title = g.Title
	t.Description = g.Notes
	if g.Due != "" {
		t.Due, _ = time.Parse(time.RFC3339, g.Due)
	}
	if g.Status == "completed" {
		t.Completed = true
	} else {
		t.Completed = false
	}

	return t
}

func ConverGTasksList(g *gtasks.TaskList) (t tasks.TasksList) {
	t.Title = g.Title
	t.GoogleID = g.Id
	return t
}
