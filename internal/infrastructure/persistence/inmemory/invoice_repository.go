package inmemory

import (
	"errors"
	"sync"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/invoice"
)

var ErrInvoiceNotFound = errors.New("invoice not found")

type InvoiceRepository struct {
	mu       sync.RWMutex
	invoices map[string]*invoice.Invoice
}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{
		mu:       sync.RWMutex{},
		invoices: make(map[string]*invoice.Invoice),
	}
}

func (r *InvoiceRepository) Save(inv *invoice.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invoices[inv.ID] = inv
	return nil
}

func (r *InvoiceRepository) FindByID(id string) (*invoice.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, ok := r.invoices[id]
	if !ok {
		return nil, ErrInvoiceNotFound
	}

	return inv, nil
}

func (r *InvoiceRepository) UpdateStatus(id string, status invoice.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inv, ok := r.invoices[id]
	if !ok {
		return ErrInvoiceNotFound
	}

	inv.Status = status
	return nil
}
