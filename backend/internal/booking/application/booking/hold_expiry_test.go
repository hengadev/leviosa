package booking

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingExpirer struct {
	calls atomic.Int32
	err   error
}

func (e *countingExpirer) ExpireElapsedHolds(context.Context) (int, error) {
	e.calls.Add(1)
	return 0, e.err
}

func TestHoldExpiryJob_TickOnceExpiresHolds(t *testing.T) {
	expirer := &countingExpirer{}
	job := newHoldExpiryJob(expirer, time.Hour)
	job.TickOnce(context.Background())
	if got := expirer.calls.Load(); got != 1 {
		t.Fatalf("expected one expiry pass, got %d", got)
	}
}

func TestHoldExpiryJob_TickFailureDoesNotStopTheJob(t *testing.T) {
	expirer := &countingExpirer{err: errors.New("database down")}
	job := newHoldExpiryJob(expirer, 10*time.Millisecond)

	done := make(chan struct{})
	go func() {
		job.Start(context.Background())
		close(done)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for expirer.calls.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	job.Stop()
	job.Stop() // safe twice
	<-done
	if got := expirer.calls.Load(); got < 3 {
		t.Fatalf("expected the job to keep ticking after a failure, got %d passes", got)
	}
}

func TestHoldExpiryJob_DefaultInterval(t *testing.T) {
	if job := newHoldExpiryJob(&countingExpirer{}, 0); job.interval != DefaultHoldExpiryInterval {
		t.Fatalf("interval = %s, want %s", job.interval, DefaultHoldExpiryInterval)
	}
}
