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
	"github.com/zeerodex/goot/pkg/timeutil"
)

const (
	TodoistAPIURL = "https://api.todoist.com/api/v1"
)

type TodoistClient struct {
	client *http.Client
	token  string
}

func NewTodoistClient(token string) apis.API {
	return &TodoistClient{
		client: http.DefaultClient,
		token:  token,
	}
}

type Task struct {
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

func (tt *Task) Task() *tasks.Task {
	var t tasks.Task
	t.TodoistID = tt.ID
	t.Title = tt.Content
	t.Description = tt.Description
	t.Due, _ = timeutil.Parse(tt.Due.Date)
	if tt.CompletedAt != "" {
		t.Completed = true
	} else {
		t.Completed = false
	}
	t.Deleted = tt.IsDeleted
	t.LastModified, _ = time.Parse(time.RFC3339, tt.UpdatedAt)
	return &t
}

func TodoistTask(t *tasks.Task) *Task {
	var tt Task
	t.TodoistID = tt.ID
	tt.Content = t.Title
	tt.Description = t.Description
	tt.Due.Date = t.Due.Format(time.RFC3339)
	if t.Completed {
		tt.CompletedAt = time.Now().Format(time.RFC3339)
	} else {
		tt.CompletedAt = ""
	}
	tt.IsDeleted = t.Deleted
	tt.UpdatedAt = t.LastModified.Format(time.RFC3339)
	return &tt
}

type response struct {
	Tasks       []Task `json:"results,omitempty"`
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

type taskCU struct {
	Content     string `json:"content"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	DueDateTime string `json:"due_datetime,omitempty"`
}

func newTaskCU(task *tasks.Task) *taskCU {
	ct := &taskCU{
		Content:     task.Title,
		Description: task.Description,
	}
	if timeutil.IsOnlyDate(task.Due) {
		ct.DueDate = task.Due.Format("2006-01-02")
	} else {
		ct.DueDateTime = task.Due.Format(time.RFC3339)
	}
	return ct
}

func (c TodoistClient) CreateTask(task *tasks.Task) (*tasks.Task, error) {
	resp, err := c.makeRequest("POST", "/tasks", newTaskCU(task))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	tt := &Task{}
	err = json.NewDecoder(resp.Body).Decode(tt)
	if err != nil {
		return nil, fmt.Errorf("error encoding json: %w", err)
	}

	return tt.Task(), nil
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

	var task Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("error encoding json: %w", err)
	}

	return task.Task(), nil
}

func (TodoistClient) GetAllTasksWithDeleted() (_ tasks.Tasks, _ error) {
	panic("not implemented") // TODO: Implement
}

func (c *TodoistClient) PatchTask(task *tasks.Task) (*tasks.Task, error) {
	resp, err := c.makeRequest("POST", fmt.Sprintf("/tasks/%s", task.TodoistID), newTaskCU(task))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	var t Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("errror encoding json: %w", err)
	}

	return t.Task(), nil
}

func (c *TodoistClient) SetTaskCompleted(id string, completed bool) error {
	method := "close"
	if completed {
		method = "reopen"
	}
	resp, err := c.makeRequest("POST", fmt.Sprintf("/tasks/%s/%s", id, method), nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return err
	}

	return nil
}

func (c *TodoistClient) DeleteTaskByID(id string) error {
	resp, err := c.makeRequest("DELETE", fmt.Sprintf("/tasks/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if err := apis.HandleResponseStatusCode(resp.StatusCode); err != nil {
		return err
	}

	return nil
}
