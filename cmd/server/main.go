package main

import (
	"github.com/rcarvalho-pb/payment_system-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/eventbus"
)

func main() {
	bus := eventbus.NewInMemoryBus()

	var invoiceRepo invoice.Repository
	var paymentRepo worker.PaymentRepository

	paymentProcessor := &worker.PaymentProcessor{
		Repo:     paymentRepo,
		EventBus: bus,
	}

	invoiceHandler := invoice.PaymentEventHandler{
		Repo: invoiceRepo,
	}

	bus.Subscribe(
		"PaymentRequested",
		paymentProcessor.Handle,
	)

	bus.Subscribe(
		"PaymentSucceeded",
		invoiceHandler.Handle,
	)

	bus.Subscribe(
		"PaymentFailed",
		invoiceHandler.Handle,
	)
}
