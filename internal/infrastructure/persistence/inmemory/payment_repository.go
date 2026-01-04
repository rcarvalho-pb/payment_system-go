package inmemory

import (
	"errors"
	"maps"
	"sync"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepository struct {
	mu              sync.RWMutex
	payments        map[string]*payment.Payment
	idempotencyKeys map[string]string
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{
		mu:              sync.RWMutex{},
		payments:        make(map[string]*payment.Payment),
		idempotencyKeys: make(map[string]string),
	}
}

func (r *PaymentRepository) Save(p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payments[p.ID] = p
	r.idempotencyKeys[p.IdempotencyKey] = p.ID
	return nil
}

func (r *PaymentRepository) SaveIfNotExist(p *payment.Payment) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.idempotencyKeys[p.IdempotencyKey]; exists {
		return false, nil
	}

	r.payments[p.ID] = p
	r.idempotencyKeys[p.IdempotencyKey] = p.ID

	return true, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(key string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paymentID, ok := r.idempotencyKeys[key]
	if !ok {
		return nil, ErrPaymentNotFound
	}
	p, ok := r.payments[paymentID]
	if !ok {
		return nil, ErrPaymentNotFound
	}

	return p, nil
}

func (r *PaymentRepository) UpdateStatus(id string, paymentStatus payment.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return ErrPaymentNotFound
	}

	p.Status = paymentStatus
	return nil
}

func (r *PaymentRepository) Payments() map[string]*payment.Payment {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return maps.Clone(r.payments)
}
