package event

type PaymentRequestPayload struct {
	InvoiceID string
	Amount    int64
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
