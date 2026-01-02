package payment

type Status string

const (
	StatusCreated    Status = "CREATED"
	StatusProcessing Status = "PROCESSING"
	StatusSuccess    Status = "SUCCESS"
	StatusFailed     Status = "FAILED"
)

type Payment struct {
	ID             string
	InvoiceID      string
	Attempt        int
	Status         Status
	IdempotencyKey string
}
