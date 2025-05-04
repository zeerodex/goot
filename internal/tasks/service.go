package tasks

type TaskService interface {
	Create(task string, description string) error
	All() (Tasks, error)
	GetByID(id int) (*Task, error)
	Update(fields ...string) (*Task, error)
	DeleteByID(id int) error
}

type taskService struct {
	r TaskRepository
}

func (s *taskService) Create(task string, description string) error {
	return s.r.Create(task, description)
}

func (s *taskService) All() (Tasks, error) {
	return s.r.GetAll()
}

func (s *taskService) GetByID(id int) (_ *Task, _ error) {
	panic("not implemented") // TODO: Implement
}

func (s *taskService) Update(fields ...string) (_ *Task, _ error) {
	panic("not implemented") // TODO: Implement
}

func (s *taskService) DeleteByID(id int) error {
	panic("not implemented") // TODO: Implement
}

func NewTaskService(repository TaskRepository) TaskService {
	return &taskService{r: repository}
}
