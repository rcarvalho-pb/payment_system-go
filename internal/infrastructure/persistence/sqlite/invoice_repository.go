package sqlite

import (
	"database/sql"
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/invoice"
)

var ErrInvoiceNotFound = errors.New("invoice not found")

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{
		db: db,
	}
}

func (r *InvoiceRepository) Save(inv *invoice.Invoice) error {
	_, err := r.db.Exec(
		`INSERT INTO invoices (id, amount, status)
		 VALUES (?, ?, ?)`,
		inv.ID,
		inv.Amount,
		string(inv.Status),
	)
	return err
}

func (r *InvoiceRepository) FindByID(id string) (*invoice.Invoice, error) {
	row := r.db.QueryRow(
		`SELECT id, amount, status
		 FROM invoices
		 WHERE id = ?`,
		id,
	)

	var inv invoice.Invoice
	var status string

	if err := row.Scan(&inv.ID, &inv.Amount, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}

	inv.Status = invoice.Status(status)
	return &inv, nil
}

func (r *InvoiceRepository) UpdateStatus(id string, status invoice.Status) error {
	res, err := r.db.Exec(
		`UPDATE invoices
		 SET status = ?
		 WHERE id = ?`,
		string(status),
		id,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrInvoiceNotFound
	}

	return nil
}
