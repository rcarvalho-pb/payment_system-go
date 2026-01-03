package event

type PaymentRequestPayload struct {
	InvoiceID string
	Amount    int64
	Attempt   int
}

type PaymentSucceededPayload struct {
	InvoiceID string
	PaymentID string
}

type PaymentFailedPayload struct {
	InvoiceID string
	PaymentID string
	Retryable bool
	Reason    string
}
