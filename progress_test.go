package progress

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"sync"
	"testing"
	"time"
)

func TestWaitForSignal(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*progress) context.Context
		cleanup         []func()
		defaultDeadline time.Duration
		wantErr         error
	}{
		{
			name: "context without deadline, signal not received",
			setup: func(p *progress) context.Context {
				return context.Background()
			},
			cleanup:         nil,
			defaultDeadline: 1 * time.Second,
			wantErr:         context.DeadlineExceeded,
		},
		{
			name: "context with deadline, signal not received",
			setup: func(p *progress) context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 20*time.Millisecond)
				return ctx
			},
			cleanup: nil,
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "signal received before context deadline",
			setup: func(p *progress) context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
				go func() {
					time.Sleep(10 * time.Millisecond)
					p.startChan <- struct{}{}
				}()
				return ctx
			},
			cleanup: nil,
			wantErr: nil,
		},
		{
			name: "context cancelled",
			setup: func(p *progress) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			cleanup: nil,
			wantErr: context.Canceled,
		},
		{
			name: "cleanup functions are executed",
			setup: func(p *progress) context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
				return ctx
			},
			cleanup: []func(){
				func() { /* simulate cleanup */ },
				func() { /* another cleanup */ },
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl := 1 * time.Minute
			if test.defaultDeadline != 0 {
				dl = test.defaultDeadline
			}
			p := &progress{
				ID:              "test-id",
				State:           "initial",
				Details:         make(map[string]any),
				Mutex:           sync.Mutex{},
				startChan:       make(chan struct{}, 1),
				Clients:         make(map[*websocket.Conn]bool),
				defaultDeadline: dl,
			}

			ctx := test.setup(p)
			err := p.WaitForSignal(ctx, test.cleanup...)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("WaitForSignal() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
