package common

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Event interface {
	Name() string
	Timestamp() time.Time
}

type EventBase struct {
	name string
	ts   time.Time
}

func NewEventBase(name string, ts time.Time) EventBase {

	if ts.IsZero() {
		ts = time.Now()
	}

	return EventBase{
		name: name,
		ts:   ts,
	}
}

func (eb *EventBase) Name() string {
	return eb.name
}

func (eb *EventBase) Timestamp() time.Time {
	return eb.ts
}

type EventBus interface {
	Send(ev Event) error
	Subscribe(s Subscriber) error
	Unsubscribe(s Subscriber) error
	Close() error
}

type Subscriber interface {
	OnEvent(e Event) error
}

type eventBus struct {
	sync.Mutex
	ch     chan Event
	logger *zap.Logger
	subs   []Subscriber
}

func NewEventBus(logger *zap.Logger) (EventBus, error) {
	eb := &eventBus{
		ch:     make(chan Event, 100),
		logger: logger,
	}
	go eb.consume()
	return eb, nil
}

func (eb *eventBus) Send(e Event) error {
	eb.Lock()
	defer eb.Unlock()
	if eb.ch != nil {
		eb.ch <- e
	}
	return nil
}

func (eb *eventBus) consume() {
	for e := range eb.ch {
		eb.logger.Debug("Received bus event", zap.String("type", e.Name()))
		for _, sub := range eb.subs {
			err := sub.OnEvent(e)
			if err != nil {
				eb.logger.Error("Error sending event", zap.Error(err))
			}
		}
	}
}
func (eb *eventBus) Subscribe(s Subscriber) error {
	eb.Lock()
	defer eb.Unlock()
	for _, sub := range eb.subs {
		if sub == s {
			return nil
		}
	}

	eb.subs = append(eb.subs, s)
	return nil
}

func (eb *eventBus) Unsubscribe(s Subscriber) error {
	eb.Lock()
	defer eb.Unlock()
	for i, sub := range eb.subs {
		if sub == s {
			eb.subs = append(eb.subs[0:i], eb.subs[i:]...)
			break
		}
	}

	return nil
}

func (eb *eventBus) Close() error {
	eb.Lock()
	defer eb.Unlock()
	if eb.ch != nil {
		close(eb.ch)
		eb.ch = nil
	}
	return nil
}
