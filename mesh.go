package gotalk

import (
	"fmt"
	//"net/http"
	//"encoding/json"
	//"math/rand"
	"time"
)

var (
	//MeshDebug Debug Switch make true to see debug Logs
	MeshDebug = false
	//MeshMaxFailCount Threshold after which tries are not made to connect to a host
	MeshMaxFailCount = 4
)

func gtLog(pattern string, message ...interface{}) string {
	s := fmt.Sprintf("mesh.go\t"+pattern, message...)
	if MeshDebug {
		fmt.Println(s)
	}
	return s
}

//MeshHostStatus Store the fail count and last change details
type MeshHostStatus struct {
	FailCount  int
	LastChange time.Time
}

//MeshHostMap Map of Host details
type MeshHostMap struct {
	HostName   string
	HostStatus MeshHostStatus
}

//MeshService Map of Service Details
type MeshService struct {
	ServiceName string
	HostMap     []MeshHostMap
}

//UpdateStatus Change Host Status for the service
func (gm GtMan) UpdateStatus(serviceName string, url string, status bool) {
	gtLog("MeshService.UpdateStatus(serviceName=%s, url=%s ,status=%t)", serviceName, url, status)
	gtLog("Before %#v", gm.ServiceMap)
	if len(gm.ServiceMap) == 0 {
		mhp := MeshHostMap{
			HostName:   url,
			HostStatus: MeshHostStatus{FailCount: 0, LastChange: time.Now()},
		}
		var ms MeshService
		ms.ServiceName = serviceName
		ms.HostMap = append(ms.HostMap, mhp)
		gm.ServiceMap = append(gm.ServiceMap, ms)
	}
	gtLog("After %#v", gm.ServiceMap)
	/*if sm[serviceName] == nil {
		hostStatus :=
		var hostMap HostMap
		hostMap = map[string]HostStatus{url: hostStatus}

		serviceMap = map[string]HostMap{serviceName: hostMap}
		sm[serviceName]

	}
	hostMap := sm[serviceName]
	if hostMap != nil {
		hostStatus := hostMap[url]
		gtLog("%#v", hostStatus)
		//[url].LastChange = time.Now()
		if status {
			//sm[serviceName][url].FailCount = 0
		} else {
			//sm[serviceName][url].FailCount++
		}
	} else {

		hostMap := make(map[string]HostStatus)
		hostMap[url] = hostStatus
		sm[serviceName] = hostMap
		gtLog("ServiceMap.UpdateStatus\t%s being added to ServiceMap for %s", url, serviceName)
	}*/
}

//GtMan Structure to hold the Manager
type GtMan struct {
	Name            string
	TCPServerURL    string
	TCPServer       *Server
	WSServerURL     string
	WSServer        *WebSocketServer
	ServiceMap      []MeshService
	Handlers        *Handlers
	OnWSPeerConnect SockHandler
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager(d bool, name string, serverName string) GtMan {
	MeshDebug = d
	gtLog("NewManager\td:%t\tname:%s\tserverName:%s", d, name, serverName)

	Name := name
	TCPServerURL := serverName
	var TCPServer *Server
	WSServerURL := serverName
	var WSServer *WebSocketServer
	var ServiceMap []MeshService
	Handlers := NewHandlers()
	var OnWSPeerConnect SockHandler

	return GtMan{
		Name:            Name,
		TCPServerURL:    TCPServerURL,
		TCPServer:       TCPServer,
		WSServerURL:     WSServerURL,
		WSServer:        WSServer,
		ServiceMap:      ServiceMap,
		Handlers:        Handlers,
		OnWSPeerConnect: OnWSPeerConnect,
	}
}
func addService(gm GtMan, op string, fn BufferReqHandler, serverName string) {
	gtLog("addService\tgm.Name:%#v\top:%s\tfn:%#v\tserverName:%s", gm.Name, op, fn, serverName)
	gm.Handlers.HandleBufferRequest(op, fn)
	gm.UpdateStatus(op, serverName, true)
}

//AddService Add a new Handler in this server
func (gm GtMan) AddService(op string, fn BufferReqHandler) {
	gtLog("GtMan.AddService\top:%s\tfn:%#v", op, fn)
	addService(gm, op, fn, gm.TCPServerURL)
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
	gtLog("GtMan.StartTCPServer\turl:%s", gm.TCPServerURL)

	s, err := Listen("tcp", gm.TCPServerURL)
	if err != nil {
		gtLog("GtMan.StartTCPServer\tListen Failed Error:%s", err.Error())
		return nil, err
	}
	gm.TCPServer = s
	gm.TCPServer.Handlers = gm.Handlers
	gm.TCPServer.Limits = NewLimits(0, 0)
	gm.TCPServer.Limits.SetReadTimeout(16 * time.Second)
	//gm.TCPServer.OnHeartbeat = func(load int, t time.Time) {		gtLog("GtMan.StartTCPServer\tgm.TCPServer.OnHeartbeat = func(load=%d, t=%s)\ts.Addr():%s", load, t, gm.TCPServer.Addr())	}

	go gm.TCPServer.Accept()

	addService(gm, "echo", echoHandler, gm.TCPServerURL)

	return gm.TCPServer, err
}

func echoHandler(s *Sock, name string, in []byte) ([]byte, error) {
	gtLog("echoHandler\ts.Addr():%s\tname:%s,in:%s", s.Addr(), name, string(in))
	return in, nil
}

//Request send Request for Service
func (gm GtMan) Request(serviceName string, param []byte) ([]byte, error) {
	gtLog("GtMan.Request\tserviceName:%s\tparam:%s", serviceName, string(param))
	/*hostMap := gm.ServiceMap[serviceName]
	if hostMap != nil {
		gtLog("%#v", hostMap)
		for url, hostStatus := range hostMap {
			gtLog("%s %#v", url, hostStatus)
			if hostStatus.FailCount < MeshMaxFailCount {
				conn, err := Connect("tcp", url)
				if err != nil {
					gtLog("GtMan.Request\tError:%s not responding\tError:%s", url, err.Error())
					gm.ServiceMap.UpdateStatus(serviceName, url, false)
				} else {
					//defer conn.Close()
					// As the responder has a one second timeout, set our heartbeat interval to half that time
					conn.HeartbeatInterval = 1 * time.Minute
					//conn.CloseHandler = func(s *Sock, code int) {					gtLog("%s Closed with code:%d", s.Addr(), code)				}
					res, err := conn.BufferRequest(serviceName, param)
					if err != nil {
						gtLog("GtMan.Request\tconn.BufferRequest(serviceName=%s,param=%s)\tres:%s\tError:%s", serviceName, param, res, err.Error())
					}
					return res, err
				}
				//defer conn.Close()
			} else {
				gtLog("GtMan.Request\thost blacklisted service:%s\tk:%s\thostMap:%#v", serviceName, url, hostMap)
			}
		}
	} else {
		gtLog("%#v", hostMap)
	}*/

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
			for service, hosts := range gm.ServiceMap {
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
func (gm GtMan) StartWSServer(url string, OnWSPeerConnect SockHandler, enableFileServer bool, httpHandlers http.Handler) (*WebSocketServer, error) {
	gtLog("GtMan.StartWSServer\turl:%s\tenableFileServer:%t", url, enableFileServer)
	gm.WSServerURL = url
	gm.OnWSPeerConnect = OnWSPeerConnect

	ws := WebSocketHandler()
	ws.OnAccept = gm.OnWSPeerConnect
	go func() { //[TODO]Remove this later and send err in return
		http.Handle("/gotalk/", ws)
		if enableFileServer {
			http.Handle("/", http.FileServer(http.Dir(".")))
		}

		http.ListenAndServe(gm.WSServerURL, httpHandlers)
	}()

	gm.WSServer = ws
	return ws, nil
}*/
