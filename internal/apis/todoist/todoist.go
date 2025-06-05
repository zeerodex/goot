package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	TodoistAPIURL = "https://api.todoist.com/api/v1"
)

type Client struct {
	client *http.Client
	token  string
}

func NewClient(token string) *Client {
	return &Client{
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
	Completed string `json:"completed_at,omitempty"`
	IsDeleted bool   `json:"is_deleted"`
	UpdatedAt string `json:"updated_at"`
}

type response struct {
	Tasks       []task `json:"results,omitempty"`
	Next_cursor string `json:"next_cursor,omitempty"`
}

func (c *Client) makeRequest(method string, endpoint string, data any) (*http.Response, error) {
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

func (c *Client) GetTasks() ([]task, error) {
	resp, err := c.makeRequest("GET", "/tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var response response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error encoding json: %w", err)
	}

	return response.Tasks, nil
}
