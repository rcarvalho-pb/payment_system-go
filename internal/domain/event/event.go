package event

type Type string

const (
	PaymentRequested Type = "REQUESTED"
	PaymentSucceeded Type = "SUCCEEDED"
	PaymentFailed    Type = "FAILED"
)

type Event struct {
	Type    Type
	Payload any
}
