package invoice

type Repository interface {
	Save(*Invoice) error
	FindByID(int64) (*Invoice, error)
	UpdateStatus(id string, status Status) error
}
