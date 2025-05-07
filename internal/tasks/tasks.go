package tasks

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      bool   `json:"status"`
}

type Tasks []Task
