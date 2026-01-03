package metrics

import "sync/atomic"

type Counters struct {
	PaymentsProcessed uint64
	PaymentsFailed    uint64
	PaymentsSucceeded uint64
}

func (c *Counters) IncProcessed() {
	atomic.AddUint64(&c.PaymentsProcessed, 1)
}

func (c *Counters) IncFailed() {
	atomic.AddUint64(&c.PaymentsFailed, 1)
}

func (c *Counters) IncSucceeded() {
	atomic.AddUint64(&c.PaymentsSucceeded, 1)
}
