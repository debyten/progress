package progress

import (
	"github.com/debyten/httplayer"
)

func NewServer() Server {
	s := NewService()
	return Server{
		Routes:  NewApi(s),
		Service: s,
	}
}

type Server struct {
	Routes  httplayer.Routing
	Service Service
}
