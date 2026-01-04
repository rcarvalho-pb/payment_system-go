package worker_test

import (
	"testing"

	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/persistence/inmemory"
)

type fakeRetry struct {
	scheduleFn func(event.PaymentRequestPayload)
}

func (f *fakeRetry) Schedule(payload event.PaymentRequestPayload) {
	f.scheduleFn(payload)
}

type fakeEventBus struct {
	publishFn func(event.Event) error
}

func (f *fakeEventBus) Publish(evt event.Event) error {
	return f.publishFn(evt)
}

type fakeExecutor struct {
	executeFn func() bool
}

func (f *fakeExecutor) Execute() bool {
	return f.executeFn()
}

type noopLogger struct{}

func (n *noopLogger) Info(string, map[string]any)  {}
func (n *noopLogger) Error(string, map[string]any) {}

func TestPaymentProcessor_WhenPaymentSucceeds_ShouldSavePaymentAndPublishEvent(t *testing.T) {
	// arrange
	repo := inmemory.NewPaymentRepository()

	publishedEvents := []event.Event{}
	eventBus := &fakeEventBus{
		publishFn: func(evt event.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	}

	executor := &fakeExecutor{
		executeFn: func() bool {
			return true
		},
	}

	metrics := &metrics.Counters{}
	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		EventBus: eventBus,
		Retry:    nil, // não é usado no caminho de sucesso
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-1",
			Amount:    1000,
			Attempt:   1,
		},
	}

	// act
	err := processor.Handle(evt)
	// assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.PaymentsProcessed != 1 {
		t.Errorf("expected PaymentsProcessed = 1, got %d", metrics.PaymentsProcessed)
	}

	if metrics.PaymentsSucceeded != 1 {
		t.Errorf("expected PaymentsSucceeded = 1, got %d", metrics.PaymentsSucceeded)
	}

	if metrics.PaymentsFailed != 0 {
		t.Errorf("expected PaymentsSucceeded = 0, got %d", metrics.PaymentsFailed)
	}

	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(publishedEvents))
	}

	if publishedEvents[0].Type != event.PaymentSucceeded {
		t.Errorf("expected PaymentSucceeded event")
	}

	p, err := repo.FindByIdempotencyKey("payment:inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.InvoiceID != "inv-1" {
		t.Errorf("expect invoice id inv-1, got %s", p.InvoiceID)
	}

	if p.Status != payment.StatusSuccess {
		t.Errorf("expect status SUCCESS, got %s", p.Status)
	}
}

func TestPaymentProcessor_WhenPaymentFails_ShouldPublishFailureAndScheduleRetry(t *testing.T) {
	repo := inmemory.NewPaymentRepository()
	publishedEvents := []event.Event{}
	eventBus := &fakeEventBus{
		publishFn: func(evt event.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	}

	executor := &fakeExecutor{
		executeFn: func() bool {
			return false
		},
	}

	retryCalled := false
	retry := &fakeRetry{
		scheduleFn: func(payload event.PaymentRequestPayload) {
			retryCalled = true
		},
	}

	metrics := &metrics.Counters{}
	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		EventBus: eventBus,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-1",
			Amount:    1000,
			Attempt:   1,
		},
	}

	err := processor.Handle(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.PaymentsProcessed != 1 {
		t.Errorf("expected paymentProcessed = 1, got %d", processor.Metrics.PaymentsProcessed)
	}

	if metrics.PaymentsFailed != 1 {
		t.Errorf("expected payment failed = 1, got %d", metrics.PaymentsFailed)
	}

	if metrics.PaymentsSucceeded != 0 {
		t.Errorf("expected payment succeeded = 0, got %d", metrics.PaymentsSucceeded)
	}

	if !retryCalled {
		t.Errorf("expected retry to be called")
	}

	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(publishedEvents))
	}

	if publishedEvents[0].Type != event.PaymentFailed {
		t.Errorf("expected PaymentFailed event")
	}
}

func TestPaymentProcessor_ShouldBiIdempotent_ForSameInvoice(t *testing.T) {
	repo := inmemory.NewPaymentRepository()

	executorCalls := 0
	executor := &fakeExecutor{
		executeFn: func() bool {
			executorCalls++
			return true
		},
	}

	eventBus := &fakeEventBus{
		publishFn: func(evt event.Event) error {
			return nil
		},
	}
	retry := &fakeRetry{
		scheduleFn: func(evt event.PaymentRequestPayload) {},
	}
	metrics := &metrics.Counters{}
	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		EventBus: eventBus,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-123",
			Amount:    500,
			Attempt:   1,
		},
	}

	_ = processor.Handle(evt)
	_ = processor.Handle(evt)

	if executorCalls != 1 {
		t.Errorf("expected executor to be called once, got %d", executorCalls)
	}
}
