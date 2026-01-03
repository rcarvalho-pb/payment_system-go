package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/eventbus"
	httpapi "github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/http"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/persistence/inmemory"
)

func main() {
	bus := eventbus.NewInMemoryBus()

	invoiceRepo := inmemory.NewInvoiceRepository()
	paymentRepo := inmemory.NewPaymentRepository()

	invoiceService := &invoice.Service{
		Repo:     invoiceRepo,
		EventBus: bus,
	}

	retryScheduler := &worker.RetryScheduler{
		EventBus:  bus,
		MaxRetry:  3,
		BaseDelay: time.Second,
		MaxDelay:  30 * time.Second,
	}

	logger := &logging.StdoutLogger{}
	metrics := &metrics.Counters{}
	executor := &worker.RandomPaymentExecutor{}

	paymentProcessor := &worker.PaymentProcessor{
		Repo:     paymentRepo,
		EventBus: bus,
		Retry:    retryScheduler,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	invoiceEventHandler := invoice.PaymentEventHandler{
		Repo: invoiceRepo,
	}

	bus.Subscribe(
		"PaymentRequested",
		paymentProcessor.Handle,
	)

	bus.Subscribe(
		"PaymentSucceeded",
		invoiceEventHandler.Handle,
	)

	bus.Subscribe(
		"PaymentFailed",
		invoiceEventHandler.Handle,
	)

	invoiceHandler := &httpapi.InvoiceHandler{
		Service: invoiceService,
	}

	router := httpapi.NewRouter(invoiceHandler)

	log.Println("HTTP server running on port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
