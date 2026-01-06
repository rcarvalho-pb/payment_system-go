package main

import (
	"context"
	"log"
	"os"

	// "net/http"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/eventbus"

	// httpapi "github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/http"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/outbox"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/persistence/sqlite"
)

func main() {
	log.Println(os.Getwd())
	db, err := sqlite.Open("./db/db.db")
	if err != nil {
		log.Fatal("error openning database")
	}

	invoiceRepo := sqlite.NewInvoiceRepository(db)
	paymentRepo := sqlite.NewPaymentRepository(db)
	outboxRepo := sqlite.NewOutboxRepository(db)

	bus := eventbus.NewInMemoryBus()
	outboxRecorder := &outbox.Recorder{
		Repo: outboxRepo,
	}

	// invoiceService := &invoice.Service{
	// 	Repo:     invoiceRepo,
	// 	EventBus: bus,
	// }

	retryScheduler := &worker.RetryScheduler{
		EventBus:  bus,
		MaxRetry:  3,
		BaseDelay: time.Second,
		MaxDelay:  30 * time.Second,
	}

	logger := &logging.StdoutLogger{}
	metrics := &metrics.Counters{}
	executor := &worker.RandomPaymentExecutor{}

	dispatcher := outbox.Dispatcher{
		Repo:         outboxRepo,
		EventBus:     bus,
		PollInterval: 1 * time.Second,
		BatchSize:    1024,
	}

	ctx := context.Background()

	go func() {
		dispatcher.Run(ctx)
	}()

	paymentProcessor := &worker.PaymentProcessor{
		Repo:     paymentRepo,
		Recorder: outboxRecorder,
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

	bus.Publish(event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-123",
			Amount:    100,
			Attempt:   1,
		},
	})

	time.Sleep(10 * time.Second)

	// invoiceHandler := &httpapi.InvoiceHandler{
	// 	Service: invoiceService,
	// }

	// router := httpapi.NewRouter(invoiceHandler)

	// log.Println("HTTP server running on port :8080")
	// log.Fatal(http.ListenAndServe(":8080", router))
}
