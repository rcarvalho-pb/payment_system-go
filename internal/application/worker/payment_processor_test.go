package worker_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/eventbus"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/outbox"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/persistence/inmemory"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	schema := `
	CREATE TABLE outbox_events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		payload BLOB NOT NULL,
		published INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}

	return db
}

type fakeRetry struct {
	scheduleFn func(event.PaymentRequestPayload)
}

func (f *fakeRetry) Schedule(payload event.PaymentRequestPayload) {
	f.scheduleFn(payload)
}

type fakeRecorder struct {
	recordFn func(event.Event) error
}

func (f *fakeRecorder) Record(evt event.Event) error {
	return f.recordFn(evt)
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

func repoPayments(repo *inmemory.PaymentRepository) map[string]*payment.Payment {
	return repo.Payments()
}

func TestPaymentProcessor_WhenPaymentSucceeds_ShouldSavePaymentAndPublishEvent(t *testing.T) {
	// arrange
	repo := inmemory.NewPaymentRepository()

	publishedEvents := []event.Event{}
	recorder := &fakeRecorder{
		recordFn: func(evt event.Event) error {
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
		Recorder: recorder,
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
	recorder := &fakeRecorder{
		recordFn: func(evt event.Event) error {
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
		Recorder: recorder,
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

	recorder := &fakeRecorder{
		recordFn: func(evt event.Event) error {
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
		Recorder: recorder,
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

func TestPaymentProcessor_ShouldNotCreateDuplicatePayments_WhenEventsAreConcurrent(t *testing.T) {
	repo := inmemory.NewPaymentRepository()
	executorCalls := 0
	executor := &fakeExecutor{
		executeFn: func() bool {
			executorCalls++
			return true
		},
	}
	publishedEvents := []event.Event{}
	recorder := &fakeRecorder{
		recordFn: func(evt event.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	}
	retry := &fakeRetry{
		scheduleFn: func(payload event.PaymentRequestPayload) {
		},
	}

	metrics := &metrics.Counters{}
	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		Recorder: recorder,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-race",
			Amount:    1000,
			Attempt:   1,
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = processor.Handle(evt)
	}()

	go func() {
		defer wg.Done()
		_ = processor.Handle(evt)
	}()

	wg.Wait()

	payments := repoPayments(repo)

	if len(payments) != 1 {
		t.Fatalf("expected exactly 1 payment, but got %d (race condition detected)", len(payments))
	}
}

func TestPaymentProcessor_ShouldEmitCorrectMetrics_OnSuccess(t *testing.T) {
	repo := inmemory.NewPaymentRepository()

	executor := &fakeExecutor{
		executeFn: func() bool {
			return true
		},
	}

	recorder := &fakeRecorder{
		recordFn: func(event.Event) error {
			return nil
		},
	}

	retry := &fakeRetry{
		scheduleFn: func(event.PaymentRequestPayload) {
		},
	}

	metrics := &metrics.Counters{}

	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		Recorder: recorder,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-metrics-ok",
			Amount:    1000,
			Attempt:   1,
		},
	}

	err := processor.Handle(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.PaymentsProcessed != 1 {
		t.Errorf("expected PaymentProcessed = 1, got %d", metrics.PaymentsProcessed)
	}

	if metrics.PaymentsSucceeded != 1 {
		t.Errorf("expected PaymentSucceeded = 1, got %d", metrics.PaymentsSucceeded)
	}

	if metrics.PaymentsFailed != 0 {
		t.Errorf("expected PaymentFailed = 0, got %d", metrics.PaymentsFailed)
	}
}

func TestPaymentProcessor_ShouldEmitCorrectMetrics_OnFailure(t *testing.T) {
	repo := inmemory.NewPaymentRepository()

	executor := &fakeExecutor{
		executeFn: func() bool {
			return false
		},
	}

	eventBus := &fakeRecorder{
		recordFn: func(event.Event) error {
			return nil
		},
	}

	retry := &fakeRetry{
		scheduleFn: func(event.PaymentRequestPayload) {
		},
	}

	metrics := &metrics.Counters{}

	logger := &noopLogger{}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		Recorder: eventBus,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: "inv-metrics-ok",
			Amount:    1000,
			Attempt:   1,
		},
	}

	err := processor.Handle(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.PaymentsProcessed != 1 {
		t.Errorf("expected PaymentProcessed = 1, got %d", metrics.PaymentsProcessed)
	}

	if metrics.PaymentsSucceeded != 0 {
		t.Errorf("expected PaymentSucceeded = 0, got %d", metrics.PaymentsSucceeded)
	}

	if metrics.PaymentsFailed != 1 {
		t.Errorf("expected PaymentFailed = 1, got %d", metrics.PaymentsFailed)
	}
}

func TestPaymentFailureTriggersRetryAndEventuallySucceeded(t *testing.T) {
	bus := eventbus.NewInMemoryBus()
	repo := inmemory.NewPaymentRepository()

	calls := 0
	executor := &fakeExecutor{
		executeFn: func() bool {
			calls++
			return calls >= 2
		},
	}

	metrics := &metrics.Counters{}
	logger := &noopLogger{}

	retry := &worker.RetryScheduler{
		EventBus:  bus,
		MaxRetry:  3,
		BaseDelay: 1 * time.Millisecond,
		MaxDelay:  5 * time.Millisecond,
	}

	db := setupTestDB(t)

	outboxDB := outbox.NewSQLiteRepository(db)

	dispatcher := outbox.Dispatcher{
		Repo:         outboxDB,
		EventBus:     bus,
		PollInterval: 1 * time.Millisecond,
		BatchSize:    1,
	}

	ctx := context.Background()

	go func() {
		dispatcher.Run(ctx)
	}()

	recorder := outbox.Recorder{outboxDB}

	processor := &worker.PaymentProcessor{
		Repo:     repo,
		Recorder: &recorder,
		Retry:    retry,
		Logger:   logger,
		Metrics:  metrics,
		Executor: executor,
	}

	bus.Subscribe(event.PaymentRequested, processor.Handle)

	payload := event.PaymentRequestPayload{
		InvoiceID: "inv-123",
		Amount:    100,
		Attempt:   1,
	}

	err := bus.Publish(event.Event{
		Type:    event.PaymentRequested,
		Payload: payload,
	})

	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	require.Equal(t, uint64(2), metrics.PaymentsProcessed)
	require.Equal(t, uint64(1), metrics.PaymentsFailed)
	require.Equal(t, uint64(1), metrics.PaymentsSucceeded)

	p, err := repo.FindByIdempotencyKey("payment:inv-123")
	require.NoError(t, err)
	require.Equal(t, p.Status, payment.StatusSuccess)
	ctx.Done()
}
