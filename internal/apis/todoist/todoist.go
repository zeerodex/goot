package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeerodex/goot/internal/apis"
	"github.com/zeerodex/goot/internal/tasks"
)

const (
	TodoistAPIURL = "https://api.todoist.com/api/v1"
)

type TodoistClient struct {
	client *http.Client
	token  string
}

func (TodoistClient) CreateTask(_ *tasks.Task) (_ *tasks.Task, _ error) {
	panic("not implemented") // TODO: Implement
}

func (TodoistClient) GetAllLists() (_ tasks.TasksLists, _ error) {
	panic("not implemented") // TODO: Implement
}

func (TodoistClient) GetAllTasksWithDeleted() (_ tasks.Tasks, _ error) {
	panic("not implemented") // TODO: Implement
}

func (TodoistClient) PatchTask(task *tasks.Task) (_ *tasks.Task, _ error) {
	panic("not implemented") // TODO: Implement
}

func (TodoistClient) ToggleCompleted(id string, completed bool) (_ error) {
	panic("not implemented") // TODO: Implement
}

func (TodoistClient) DeleteTaskByID(id string) (_ error) {
	panic("not implemented") // TODO: Implement
}

func NewTodoistClient(token string) apis.API {
	return &TodoistClient{
		client: http.DefaultClient,
		token:  token,
	}
}

type task struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Description string `json:"description,omitempty"`
	Due         struct {
		Date string `json:"date,omitempty"`
	} `json:"due"`
	CompletedAt string `json:"completed_at,omitempty"`
	IsDeleted   bool   `json:"is_deleted"`
	UpdatedAt   string `json:"updated_at"`
}

func (tt *task) Task() *tasks.Task {
	var t tasks.Task
	t.Title = tt.Content
	t.Description = tt.Description
	t.Due, _ = time.Parse(time.RFC3339, tt.Due.Date)
	if tt.CompletedAt != "" {
		t.Completed = true
	} else {
		t.Completed = false
	}
	t.Deleted = tt.IsDeleted
	t.LastModified, _ = time.Parse(time.RFC3339, tt.UpdatedAt)
	return &t
}

type response struct {
	Tasks       []task `json:"results,omitempty"`
	Next_cursor string `json:"next_cursor,omitempty"`
}

func (c *TodoistClient) makeRequest(method string, endpoint string, data any) (*http.Response, error) {
	var reqBody io.Reader
	if data != nil {
		jsonBody, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	req, err := http.NewRequest(method, TodoistAPIURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

func (c *TodoistClient) GetAllTasks() (tasks.Tasks, error) {
	resp, err := c.makeRequest("GET", "/tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	var response response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error encoding json: %w", err)
	}

	// HACK:
	tasks := make(tasks.Tasks, len(response.Tasks))
	for i, t := range response.Tasks {
		tasks[i] = *t.Task()
	}

	return tasks, nil
}

func (c *TodoistClient) GetTaskByID(id string) (*tasks.Task, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("/tasks/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	var task task
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("error encoding json: %w", err)
	}

	return task.Task(), nil
}
