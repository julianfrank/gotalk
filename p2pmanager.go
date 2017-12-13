package gotalk

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	debug bool
)

//GtMan Structure to hold the Manager
type GtMan struct {
	localTCPServerName string
	tcpServer          *Server
	localWSServerName  string
	wsServer           *WebSocketServer

	peer *Sock

	handlers *Handlers

	onWSPeerConnect SockHandler
}

func gtLog(pattern string, message ...interface{}) string {
	s := fmt.Sprintf("p2pmanager.go:\t"+pattern, message...)
	if debug {
		fmt.Println(s)
	}
	return s
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager(d bool) GtMan {
	debug = d

	handlers := NewHandlers()
	peer := NewSock(handlers)

	return GtMan{
		localTCPServerName: "",
		localWSServerName:  "",
		handlers:           handlers,
		peer:               peer,
	}
}

//AddHandler Add a new Handler in this server
func (gm GtMan) AddHandler(op string, fn BufferReqHandler) {
	gtLog("GtMan.AddHandler op:%s", op)
	gm.handlers.HandleBufferRequest(op, fn)
}

//StartTCPServer Start TCP Server
func (gm GtMan) StartTCPServer(url string) (*Server, error) {
	gtLog("GtMan.StartTCPServer url:%s", url)
	gm.localTCPServerName = url

	s, err := Listen("tcp", gm.localTCPServerName)
	if err != nil {
		gtLog("Listen Failed Error:%s", err.Error())
		return nil, err
	}
	s.AcceptHandler = func(sock *Sock) {
		gtLog("%s Accept Started for %s", sock.Addr(), url)
	}
	s.OnHeartbeat = func(load int, t time.Time) {
		gtLog("load:%d\tt:%s\ts.Addr():%s", load, t, s.Addr())
	}
	s.Handlers = gm.handlers
	gm.tcpServer = s
	go gm.tcpServer.Accept()
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

	gm.wsServer = ws
	return ws, nil
}

//PeerConnect Connect with Peer
func (gm GtMan) PeerConnect(url string) (*Sock, error) {
	gtLog("GtMan.PeerConnect url:%s", url)

	conn, err := Connect("tcp", url)
	if err != nil {
		log.Fatal("Connect Failed:", err)
		return nil, err
	}

	conn.CloseHandler = func(s *Sock, code int) {
		gtLog("%s Closed! code:%d", s.Addr(), code)
	}

	res, err := conn.BufferRequest("echo", []byte("Testing"))
	if err != nil {
		gtLog("gm.Peer.BufferRequest(echo... res:%s\tError:%s", res, err.Error())
		return nil, err
	}

	gm.peer = conn

	defer conn.Close()

	return conn, err
}

//Request send Request for Service
func (gm GtMan) Request(serviceName string, param []byte) ([]byte, error) {
	if gm.peer != nil {
		if gm.peer.Addr() != "" {
			return gm.peer.BufferRequest(serviceName, param)
		} else {
			return nil, fmt.Errorf("gm.peer.Addr() Error: Not Connected [%s]", gm.peer.Addr())
		}
	}
	return nil, fmt.Errorf("gm.peer Error: Not Initialised")
}
