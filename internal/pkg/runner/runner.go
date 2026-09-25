package runner

import (
	"context"
	"sync"
)

type (
	Runner interface {
		// Run executes a function with a context and a cancellation signal.
		// The function should exit it's inner loop when signaled.
		// The context will remain valid until the called Shutdown function times out.
		Run(f func(ctx context.Context, closeSignal <-chan struct{}))
		// Shutdown signals all tasks to shutdown.
		// It waits until either all tasks are closed gracefuly or the context is cancelled.
		Shutdown(ctx context.Context) error
	}

	runner struct {
		ctx    context.Context
		cancel context.CancelFunc

		wg               sync.WaitGroup
		shutdownSignalCh chan struct{}
		closedCh         chan struct{}
	}
)

func New() Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &runner{
		ctx:              ctx,
		cancel:           cancel,
		shutdownSignalCh: make(chan struct{}),
		closedCh:         make(chan struct{}),
	}
}

func (s *runner) Run(f func(ctx context.Context, signal <-chan struct{})) {
	s.wg.Go(func() {
		f(s.ctx, s.shutdownSignalCh)
	})
}

func (s *runner) Shutdown(ctx context.Context) error {
	go func() {
		close(s.shutdownSignalCh)
		s.wg.Wait()
		close(s.closedCh)
	}()

	select {
	case <-s.closedCh:
		return nil
	case <-ctx.Done():
		s.cancel()
		return ctx.Err()
	}
}
