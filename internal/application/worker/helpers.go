package worker

import (
	"fmt"
	"math/rand"
	"time"
)

func generateIdempotencyKey(invoiceID string) string {
	return fmt.Sprintf("payment:%s", invoiceID)
}

func generatePaymentID() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}

func simulatePayment() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100) < 70
}
