package im

import (
	"net"
	"net/http"
	"os"
)

// HTTP server that supported graceful shutdown or restart
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	isGraceful   bool
	signalChan   chan os.Signal
	shutdownChan chan bool
}

var server *http.Server

type Handler struct {

}

func (handler *Handler) ServeHTTP(http.ResponseWriter, *http.Request)  {

}

func Start(addr string, handler http.Handler) error {
	server := &http.Server{
		Addr: addr,
		Handler: handler,
	}
	return server.ListenAndServe()
}