package gotalk

import (
	"fmt"
	//"net/http"
	//"encoding/json"
	"time"
)

var debug bool

func gtLog(pattern string, message ...interface{}) string {
	s := fmt.Sprintf("p2pmanager.go:\t"+pattern, message...)
	if debug {
		fmt.Println(s)
	}
	return s
}

//GtMan Structure to hold the Manager
type GtMan struct {
	localTCPServerName string
	tcpServer          *Server
	localWSServerName  string
	wsServer           *WebSocketServer
	services           map[string]map[string]bool //Map of [service name]array of servers providing the service
	handlers           *Handlers
	onWSPeerConnect    SockHandler
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager(d bool, serverName string) GtMan {
	debug = d
	gtLog("NewManager d:%t serverName:%s", d, serverName)

	localTCPServerName := serverName
	var tcpServer *Server
	localWSServerName := serverName
	wsServer := WebSocketHandler()
	services := make(map[string]map[string]bool)
	handlers := NewHandlers()
	var onWSPeerConnect SockHandler

	return GtMan{
		localTCPServerName: localTCPServerName,
		tcpServer:          tcpServer,
		localWSServerName:  localWSServerName,
		wsServer:           wsServer,
		services:           services,
		handlers:           handlers,
		onWSPeerConnect:    onWSPeerConnect,
	}
}

//AddService Add a new Handler in this server
func (gm GtMan) AddService(op string, fn BufferReqHandler) {
	gtLog("GtMan.AddService op:%s", op)
	gm.handlers.HandleBufferRequest(op, fn)
	if gm.services[op] == nil {
		gm.services[op] = make(map[string]bool)
	}
	gm.services[op][gm.localTCPServerName] = true
}

//StartTCPServer Start TCP Server
func (gm GtMan) StartTCPServer() (*Server, error) {
	gtLog("GtMan.StartTCPServer url:%s", gm.localTCPServerName)

	s, err := Listen("tcp", gm.localTCPServerName)
	if err != nil {
		gtLog("Listen Failed Error:%s", err.Error())
		return nil, err
	}
	gm.tcpServer = s
	// Configure limits with a read timeout of one second
	gm.tcpServer.Limits = NewLimits(0, 0)
	gm.tcpServer.Limits.SetReadTimeout(time.Second)
	gm.tcpServer.Handlers = gm.handlers

	gm.tcpServer.OnHeartbeat = func(load int, t time.Time) {
		gtLog("load:%d\tt:%s\ts.Addr():%s", load, t, gm.tcpServer.Addr())
	}

	go gm.tcpServer.Accept()
	return gm.tcpServer, err
}

//Request send Request for Service
func (gm GtMan) Request(serviceName string, param []byte) ([]byte, error) {
	gtLog("p2pmanager.go::GtMan.Request serviceName:%s\tparam:%s", serviceName, string(param))
	//Select host
	targetHosts := gm.services[serviceName]
	for url, t := range targetHosts {
		gtLog("k:%s\tv:%t", url, t)
		conn, err := Connect("tcp", url)
		if err != nil {
			gtLog("p2pmanager.go::GtMan.Request Error:%s not responding\tError:%s", url, err.Error())
		} else {
			// As the responder has a one second timeout, set our heartbeat interval to half that time
			conn.HeartbeatInterval = 500 * time.Millisecond
			conn.CloseHandler = func(s *Sock, code int) {
				gtLog("%s Closed with code:%d", conn.Addr(), code)
			}
			res, err := conn.BufferRequest(serviceName, param)
			if err != nil {
				gtLog("conn.BufferRequest(serviceName=%s, param=%s) res:%s\tError:%s", serviceName, param, res, err.Error())
			}
			defer conn.Close()
			return res, err
		}
		defer conn.Close()
	}
	return nil, fmt.Errorf("p2pmanager.go::GtMan.Request error: unable to connect to any hosts")
}

/*
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
}*/
