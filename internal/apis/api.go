package apis

import (
	"fmt"
	"net/http"

	"github.com/zeerodex/goot/internal/tasks"
)

type API interface {
	CreateTask(*tasks.Task) (*tasks.Task, error)
	GetTaskByID(id string) (*tasks.Task, error)
	GetAllTasks() (tasks.Tasks, error)
	UpdateTask(task *tasks.Task) (*tasks.Task, error)
	SetTaskCompleted(id string, completed bool) error
	DeleteTaskByID(id string) error
}

func HandleResponseStatusCode(statusCode int) error {
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		baseErr := fmt.Errorf("API request failed with status: %d", statusCode)

		switch statusCode {
		case http.StatusBadRequest:
			return fmt.Errorf("%w: client error, please check your request parameters", baseErr)
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: authentication required or invalid credentials", baseErr)
		case http.StatusForbidden:
			return fmt.Errorf("%w: access denied, insufficient permissions", baseErr)
		case http.StatusNotFound:
			return fmt.Errorf("%w: resource not found", baseErr)
		case http.StatusInternalServerError:
			return fmt.Errorf("%w: internal server error, please try again later", baseErr)
		default:
			return baseErr
		}
	}
	return nil
}
