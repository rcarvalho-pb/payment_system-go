package sqlite

import (
	"database/sql"
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Save(p *payment.Payment) error {
	_, err := r.db.Exec(
		`INSERT INTO payments
		 (id, invoice_id, attempt, status, idempotency_key)
		 VALUES (?, ?, ?, ?, ?)`,
		p.ID,
		p.InvoiceID,
		p.Attempt,
		string(p.Status),
		p.IdempotencyKey,
	)
	return err
}

func (r *PaymentRepository) SaveIfNotExist(p *payment.Payment) (bool, error) {
	res, err := r.db.Exec(
		`INSERT OR IGNORE INTO payments
		 (id, invoice_id, attempt, status, idempotency_key)
		 VALUES (?, ?, ?, ?, ?)`,
		p.ID,
		p.InvoiceID,
		p.Attempt,
		string(p.Status),
		p.IdempotencyKey,
	)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	// 0 rows = idempotency hit
	return affected == 1, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(key string) (*payment.Payment, error) {
	row := r.db.QueryRow(
		`SELECT id, invoice_id, attempt, status, idempotency_key
		 FROM payments
		 WHERE idempotency_key = ?`,
		key,
	)

	var p payment.Payment
	var status string

	if err := row.Scan(
		&p.ID,
		&p.InvoiceID,
		&p.Attempt,
		&status,
		&p.IdempotencyKey,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	p.Status = payment.Status(status)
	return &p, nil
}

func (r *PaymentRepository) UpdateStatus(id string, newStatus payment.Status) error {
	res, err := r.db.Exec(
		`UPDATE payments
		 SET status = ?
		 WHERE id = ?`,
		string(newStatus),
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
		return ErrPaymentNotFound
	}

	return nil
}
