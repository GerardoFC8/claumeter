package server

import "sync"

// Broker fan-outs a single message to N SSE subscribers. Slow subscribers
// have their message dropped rather than blocking the publisher.
type Broker struct {
	mu   sync.RWMutex
	subs map[chan []byte]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: map[chan []byte]struct{}{}}
}

func (b *Broker) Subscribe() chan []byte {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Broker) Publish(msg []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (b *Broker) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
