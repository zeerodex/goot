package tasks

type TaskService struct {
	r TaskRepository
}

func NewTaskService(repository TaskRepository) *TaskService {
	return &TaskService{r: repository}
}

func (r *TaskService) All() (Tasks, error) {
	return r.r.GetAll()
}
