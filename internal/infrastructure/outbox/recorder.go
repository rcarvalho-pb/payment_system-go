package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type Recorder struct {
	Repo Repository
}

func generateOutboxID() string {
	return fmt.Sprintf("outbox_%d", time.Now().UnixNano())
}

func (r *Recorder) Record(evt event.Event) error {
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return nil
	}

	return r.Repo.Save(OutboxEvent{
		ID:        generateOutboxID(),
		Type:      evt.Type,
		Payload:   payload,
		CreatedAt: time.Now(),
	})
}
