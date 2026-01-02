package invoice

type Repository interface {
	Save(*Invoice) error
	FindByID(string) (*Invoice, error)
	UpdateStatus(id string, status Status) error
}
