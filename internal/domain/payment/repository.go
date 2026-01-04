package payment

type Repository interface {
	Save(*Payment) error
	SaveIfNotExist(*Payment) (bool, error)
	FindByIdempotencyKey(string) (*Payment, error)
	UpdateStatus(string, Status) error
}
