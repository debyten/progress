package progress

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var defaultDeadline = 1 * time.Minute

type CtxKey int

const CtxValue CtxKey = iota

func Inject(ctx context.Context, prog Progress) context.Context {
	return context.WithValue(ctx, CtxValue, prog)
}

// FromContext return a progress instance from the supplied context. When empty, a noop implementation is returned.
func FromContext(ctx context.Context) Progress {
	p, ok := ctx.Value(CtxValue).(Progress)
	if !ok {
		return noop{}
	}
	return p
}

type noop struct{}

func (n noop) GetID() string {
	return ""
}

func (n noop) WaitForSignal(context.Context, ...func()) error {
	return nil
}

func (n noop) Update(_ string, _ ...map[string]any) {
}

// Progress represents the state of a long-running operation.
type Progress interface {
	GetID() string
	// WaitForSignal blocks until the first client connects or context is canceled.
	// A deadline must be set on the context, if none is provided, a default timeout of 1 minute will be used.
	//
	// An optional cleanup function can be provided to be executed when the context is canceled.
	WaitForSignal(ctx context.Context, cleanup ...func()) error
	Update(state string, details ...map[string]any)
}

type progress struct {
	ID        string
	State     string
	Details   map[string]any
	Mutex     sync.Mutex
	started   bool
	startChan chan struct{}
	Clients   map[*websocket.Conn]bool

	defaultDeadline time.Duration
}

// NewProgress initializes a new progress instance.
func NewProgress() Progress {
	return newProgress()
}

func newProgress() *progress {
	return &progress{
		ID:              uuid.NewString(),
		State:           "Pending",
		Clients:         make(map[*websocket.Conn]bool),
		startChan:       make(chan struct{}, 1),
		defaultDeadline: defaultDeadline,
	}
}

func (p *progress) GetID() string {
	return p.ID
}

func (p *progress) WaitForSignal(ctx context.Context, cleanup ...func()) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.defaultDeadline)
		defer cancel()
	}
	select {
	case <-ctx.Done():
		for _, cleanupFunc := range cleanup {
			cleanupFunc()
		}
		return ctx.Err()
	case <-p.startChan:
		return nil
	}
}

// Update updates the state and details of the progress and notifies all connected clients.
func (p *progress) Update(state string, details ...map[string]any) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.State = state
	if len(details) > 0 {
		p.Details = details[0]
	} else {
		p.Details = nil
	}
	p.notifyClients()
}

// AddClient adds a websocket connection to the list of clients.
func (p *progress) AddClient(conn *websocket.Conn) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if !p.started {
		p.started = true
		p.startChan <- struct{}{}
	}

	p.Clients[conn] = true
}

// RemoveClient removes a websocket connection from the list of clients.
func (p *progress) RemoveClient(conn *websocket.Conn) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	delete(p.Clients, conn)
	_ = conn.Close()
}

// notifyClients sends the current state and details to all connected clients.
func (p *progress) notifyClients() {
	message := map[string]any{
		"state":   p.State,
		"details": p.Details,
	}

	for conn := range p.Clients {
		if err := conn.WriteJSON(message); err != nil {
			p.RemoveClient(conn)
		}
	}
}
