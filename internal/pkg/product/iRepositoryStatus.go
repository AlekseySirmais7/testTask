package product

type IRepositoryStatus interface {
	CreateStatus(taskId int, status string) error
	SetStatus(taskId int, status string) error
	GetStatus(taskId int) (string, error)
	CloseProduct() error
}
