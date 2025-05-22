package progress

import (
	"context"
	"github.com/gorilla/websocket"
	"sync"
)

type Service interface {
	// CreateProgress creates and stores a new progress instance.
	CreateProgress() Progress
	// GetProgress retrieves a progress instance by ID.
	GetProgress(id string) (*progress, bool)
	// DeleteProgress removes a progress instance by ID.
	DeleteProgress(id string)
	// Stream progress events to ws connection
	Stream(ctx context.Context, ws *websocket.Conn, id string)
}

// manager manages multiple progress instances.
type manager struct {
	progressMap map[string]*progress
	Mutex       sync.Mutex
}

func NewService() Service {
	return &manager{
		progressMap: make(map[string]*progress),
	}
}

func (pm *manager) CreateProgress() Progress {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	progress := newProgress()
	pm.progressMap[progress.ID] = progress
	return progress
}

func (pm *manager) GetProgress(id string) (*progress, bool) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	progress, exists := pm.progressMap[id]
	return progress, exists
}

func (pm *manager) DeleteProgress(id string) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	delete(pm.progressMap, id)
}

func (pm *manager) Stream(ctx context.Context, ws *websocket.Conn, id string) {
	p, exists := pm.GetProgress(id)
	if !exists {
		_ = ws.Close()
		return
	}
	p.AddClient(ws)
	<-ctx.Done()
	p.RemoveClient(ws)
}
