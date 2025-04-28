package tasks

type Task struct {
	ID          int
	Task        string
	Description string
	Status      bool
}

type Tasks []Task
