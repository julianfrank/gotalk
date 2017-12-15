package gotalk

import (
	"fmt"
	//"net/http"
	//"encoding/json"
	//"math/rand"
	"time"
)

var debug bool

func gtLog(pattern string, message ...interface{}) string {
	s := fmt.Sprintf("mesh.go\t"+pattern, message...)
	if debug {
		fmt.Println(s)
	}
	return s
}

type hostMap map[string]bool
type serviceMap map[string]hostMap

//GtMan Structure to hold the Manager
type GtMan struct {
	Name               string
	localTCPServerName string
	tcpServer          *Server
	localWSServerName  string
	wsServer           *WebSocketServer
	serviceMap         serviceMap
	handlers           *Handlers
	onWSPeerConnect    SockHandler
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager(d bool, name string, serverName string) GtMan {
	debug = d
	gtLog("NewManager\td:%t\tname:%s\tserverName:%s", d, name, serverName)

	Name := name
	localTCPServerName := serverName
	var tcpServer *Server
	localWSServerName := serverName
	var wsServer *WebSocketServer
	serviceMap := make(serviceMap)
	handlers := NewHandlers()
	var onWSPeerConnect SockHandler

	return GtMan{
		Name:               Name,
		localTCPServerName: localTCPServerName,
		tcpServer:          tcpServer,
		localWSServerName:  localWSServerName,
		wsServer:           wsServer,
		serviceMap:         serviceMap,
		handlers:           handlers,
		onWSPeerConnect:    onWSPeerConnect,
	}
}
func addService(gm GtMan, op string, fn BufferReqHandler, serverName string) {
	gtLog("addService\tgm.Name:%#v\top:%s\tfn:%#v\tserverName:%s", gm.Name, op, fn, serverName)
	gm.handlers.HandleBufferRequest(op, fn)
	if gm.serviceMap[op] == nil {
		gm.serviceMap[op] = make(map[string]bool)
	}
	gm.serviceMap[op][serverName] = true
}

//AddService Add a new Handler in this server
func (gm GtMan) AddService(op string, fn BufferReqHandler, serverName string) {
	gtLog("GtMan.AddService\top:%s\tfn:%#v\tserverName:%s", op, fn, serverName)
	addService(gm, op, fn, serverName)
}

//AddPeer Add a New Peer in the Mesh
func (gm GtMan) AddPeer(peerName string) error {
	gtLog("GtMan.AddPeer\tpeerName:%s", peerName)
	conn, err := Connect("tcp", peerName)
	if err != nil {
		return fmt.Errorf(gtLog("GtMan.AddPeer\tError:%s not responding\tError:%s", peerName, err.Error()))
	}
	addService(gm, "echo", echoHandler, peerName)
	defer conn.Close()
	return nil
}

//StartTCPServer Start TCP Server
func (gm GtMan) StartTCPServer() (*Server, error) {
	gtLog("GtMan.StartTCPServer\turl:%s", gm.localTCPServerName)

	s, err := Listen("tcp", gm.localTCPServerName)
	if err != nil {
		gtLog("GtMan.StartTCPServer\tListen Failed Error:%s", err.Error())
		return nil, err
	}
	gm.tcpServer = s
	gm.tcpServer.Handlers = gm.handlers
	gm.tcpServer.Limits = NewLimits(0, 0)
	gm.tcpServer.Limits.SetReadTimeout(16 * time.Second)
	//gm.tcpServer.OnHeartbeat = func(load int, t time.Time) {		gtLog("GtMan.StartTCPServer\tgm.tcpServer.OnHeartbeat = func(load=%d, t=%s)\ts.Addr():%s", load, t, gm.tcpServer.Addr())	}

	go gm.tcpServer.Accept()

	addService(gm, "echo", echoHandler, gm.localTCPServerName)

	return gm.tcpServer, err
}
func echoHandler(s *Sock, name string, in []byte) ([]byte, error) {
	gtLog("echoHandler\ts.Addr():%s\tname:%s,in:%s", s.Addr(), name, string(in))
	return in, nil
}

//Request send Request for Service
func (gm GtMan) Request(serviceName string, param []byte) ([]byte, error) {
	gtLog("GtMan.Request\tserviceName:%s\tparam:%s", serviceName, string(param))
	//Select host
	targetHosts := gm.serviceMap[serviceName]
	for url, t := range targetHosts {
		if t {
			conn, err := Connect("tcp", url)
			if err != nil {
				gtLog("GtMan.Request\tError:%s not responding\tError:%s", url, err.Error())
				gm.serviceMap[serviceName][url] = false
			} else {
				defer conn.Close()
				// As the responder has a one second timeout, set our heartbeat interval to half that time
				conn.HeartbeatInterval = 1 * time.Minute
				//conn.CloseHandler = func(s *Sock, code int) {					gtLog("%s Closed with code:%d", s.Addr(), code)				}
				res, err := conn.BufferRequest(serviceName, param)
				if err != nil {
					gtLog("GtMan.Request\tconn.BufferRequest(serviceName=%s,param=%s)\tres:%s\tError:%s", serviceName, param, res, err.Error())
				}
				return res, err
			}
			defer conn.Close()
		} else {
			gtLog("GtMan.Request\thost blacklisted service:%s\tk:%s\tv:%t", serviceName, url, t)
		}
	}
	return nil, fmt.Errorf("GtMan.Request\terror: unable to connect to any hosts")
}

/*
//StartHealthChecker Check all hosts and mark health
func (gm GtMan) StartHealthChecker(preferedDelay time.Duration) {
	gtLog("GtMan.StartHealthChecker\tStarting with preferedDelay:%s", preferedDelay)
	ticker := time.NewTicker(preferedDelay)
	go func() {
		for t := range ticker.C {
			gtLog("GtMan.StartHealthchecker\tTick at %+v", t)
			for service, hosts := range gm.serviceMap {
				for host, status := range hosts {
					gtLog("GtMan.StartHealthChecker\tservice:%s\thost:%s\tstatus:%t", service, host, status)
				}
			}
		}
	}()
	//defer func() {
		//time.Sleep(25 * time.Minute)
		//gtLog("GtMan.StartHealthChecker\tStopped")
		//ticker.Stop()
	//}()
}

//StartWSServer Start WebSocket Server
//This is a Blocking Service
func (gm GtMan) StartWSServer(url string, onWSPeerConnect SockHandler, enableFileServer bool, httpHandlers http.Handler) (*WebSocketServer, error) {
	gtLog("GtMan.StartWSServer\turl:%s\tenableFileServer:%t", url, enableFileServer)
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
