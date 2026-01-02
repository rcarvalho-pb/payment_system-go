package event

type Type string

const (
	PaymentRequest   Type = "REQUEST"
	PaymentSucceeded Type = "SUCCEEDED"
	PaymentFailed    Type = "FAILED"
)

type Event struct {
	Type    Type
	Payload any
}
