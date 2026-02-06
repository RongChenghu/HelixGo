package fake

import (
	"context"
	"sync"

	"helix-bot/internal/adapter/telegram"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// FakeUpdateSource is a test-only implementation of ports.UpdateSource.
// It sequentially pushes pre-defined updates to the out channel, respecting
// ctx.Done() and Stop(), and applies a Deduper so duplicate UpdateID values
// are not processed twice.
type FakeUpdateSource struct {
	Updates []types.BotUpdate

	dedup   *telegram.Deduper
	stop    chan struct{}
	stopped chan struct{}
	mu      sync.Mutex
}

// Start starts a goroutine that emits Updates to out in order.
func (f *FakeUpdateSource) Start(ctx context.Context, out chan<- types.BotUpdate) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stop != nil {
		// already started; no-op
		return nil
	}
	f.stop = make(chan struct{})
	f.stopped = make(chan struct{})
	if f.dedup == nil {
		// default window (same as telegram source)
		f.dedup = telegram.NewDeduper(0)
	}

	go func() {
		defer close(out)
		defer close(f.stopped)

		for _, u := range f.Updates {
			select {
			case <-ctx.Done():
				return
			case <-f.stop:
				return
			default:
			}

			if f.dedup != nil && f.dedup.Seen(u.UpdateID) {
				continue
			}

			select {
			case out <- u:
			case <-ctx.Done():
				return
			case <-f.stop:
				return
			}
		}
	}()
	return nil
}

// Stop signals the source to stop and waits for the goroutine to finish.
func (f *FakeUpdateSource) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stop == nil {
		return nil
	}
	select {
	case <-f.stop:
		// already closed
	default:
		close(f.stop)
	}
	if f.stopped != nil {
		<-f.stopped
	}
	return nil
}

var _ ports.UpdateSource = (*FakeUpdateSource)(nil)
