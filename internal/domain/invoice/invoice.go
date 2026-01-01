package invoice

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusProcessing Status = "PROCESSING"
	StatusPaid       Status = "PAID"
	StatusFailed     Status = "FAILED"
	StatusCanceled   Status = "CANCELED"
)

type Invoice struct {
	ID     string
	Amount int64
	Status Status
}
