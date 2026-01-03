package worker

import (
	"fmt"
	"time"
)

func generateIdempotencyKey(invoiceID string) string {
	return fmt.Sprintf("payment:%s", invoiceID)
}

func generatePaymentID() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}
