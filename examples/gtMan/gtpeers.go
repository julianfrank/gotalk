package main

import (
	"fmt"
	"github.com/julianfrank/gotalk"
	"log"
	"time"
)

func main() {
	xURL, yURL := "localhost:9090", "localhost:9092"
	//startTime := time.Now()
	log.Println("Starting Peer x")
	x := gotalk.NewManager(true, "x", xURL)
	x.AddService("echo2", echo2Handler)
	x.StartTCPServer()

	log.Println("Starting Peer y")
	y := gotalk.NewManager(true, "y", yURL)
	y.StartTCPServer()
	x.AddPeer(yURL)

	for i := 0; i < 2; {
		r, err := x.Request("echo", []byte(fmt.Sprintf("Hello %d", i)))
		log.Println("echo Response", string(r), err)
		time.Sleep(time.Second)
		i++
	}

}

func onWSConnect(s *gotalk.Sock) {
	s.Notify("hello", "world")
	log.Println("onWSConnect", s.Addr())
}

func echo2Handler(s *gotalk.Sock, name string, in []byte) ([]byte, error) {
	log.Println("echo2Handler\ts.Addr():%s\tname:%s,in:%s", s.Addr(), name, string(in))
	return []byte(string(in) + string(in)), nil
}
