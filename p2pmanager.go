package gotalk

import (
	"fmt"
	"log"
	"net/http"
)

var (
	debug bool
)

//GtMan Structure to hold the Manager
type GtMan struct {
	localTCPServerName string
	localWSServerName  string

	handlers *Handlers

	onWSPeerConnect SockHandler
}

func gtLog(pattern string, message ...interface{}) string {
	s := fmt.Sprintf(pattern, message...)
	if debug {
		log.Println(s)
	}
	return s
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager(d bool) GtMan {
	debug = d

	handlers := NewHandlers()
	return GtMan{
		localTCPServerName: "",
		localWSServerName:  "",
		handlers:           handlers,
	}
}

//AddHandler Add a new Handler in this server
func (gm GtMan) AddHandler(op string, fn interface{}) error {
	gtLog("GtMan.AddHandler op:%s", op)
	gm.handlers.Handle(op, fn)
	return nil
}

//StartTCPServer Start TCP Server
func (gm GtMan) StartTCPServer(url string) (*Server, error) {
	gtLog("GtMan.StartTCPServer url:%s", url)
	gm.localTCPServerName = url

	s, err := Listen("tcp", gm.localTCPServerName)
	s.Handlers = gm.handlers
	go s.Accept()
	return s, err
}

//StartWSServer Start WebSocket Server
//This is a Blocking Service
func (gm GtMan) StartWSServer(url string, onWSPeerConnect SockHandler, enableFileServer bool, httpHandlers http.Handler) (*WebSocketServer, error) {
	gtLog("GtMan.StartWSServer url:%s\tenableFileServer:%t", url, enableFileServer)
	gm.localWSServerName = url
	gm.onWSPeerConnect = onWSPeerConnect

	ws := WebSocketHandler()
	ws.OnAccept = gm.onWSPeerConnect
	go func() { //[TODO]Remove this later and send err in return
		http.Handle("/gotalk/", ws)
		if enableFileServer {
			http.Handle("/", http.FileServer(http.Dir(".")))
		}

		http.ListenAndServe(gm.localWSServerName, httpHandlers)
	}()

	return ws, nil
}

//PeerConnect Connect with Peer
func (gm GtMan) PeerConnect(url string) (*Sock, error) {
	gtLog("GtMan.PeerConnect url:%s", url)

	s, err := Connect("tcp", url)
	return s, err
}
