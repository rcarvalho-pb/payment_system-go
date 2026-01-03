package worker

import (
	"math/rand"
)

type RandomPaymentExecutor struct{}

func (r *RandomPaymentExecutor) Execute() bool {
	return rand.Intn(100) < 70
}
