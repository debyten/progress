package progress

import (
	"context"
	"github.com/debyten/apierr"
	"github.com/debyten/httplayer"
	"github.com/gorilla/websocket"
	"net/http"
)

// PathParamFunc is a function variable that extracts path parameters from HTTP requests.
// It can be customized to support different router implementations.
// By default, it uses the standard http.Request.PathValue method.
var PathParamFunc = func(r *http.Request, name string) string {
	return r.PathValue(name)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewApi(pm Service) httplayer.Routing {
	return api{pm: pm}
}

type api struct {
	pm Service
}

func (a api) Routes(with *httplayer.RoutingDefinition) []httplayer.Route {
	return with.
		Add(http.MethodGet, "/api/v1/progress/{id}", a.streamProgress).
		Done()
}

func (a api) streamProgress(w http.ResponseWriter, r *http.Request) {
	id := PathParamFunc(r, "id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		apierr.Handle(err, w)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})
	go a.pm.Stream(ctx, conn, id)
}
