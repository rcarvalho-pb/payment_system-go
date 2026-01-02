package event

type PaymentRequestPayload struct {
	InvoiceID string
	Amount    int64
}
