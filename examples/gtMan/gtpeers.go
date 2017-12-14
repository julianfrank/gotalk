package main

import (
	"github.com/julianfrank/gotalk"
	"log"
	"time"
)

func main() {
	//startTime := time.Now()
	log.Println("Starting Peer x")
	x := gotalk.NewManager(true, "localhost:9090")
	x.StartHealthChecker(time.Millisecond * 1111)
	x.AddService("echo", echo)
	x.StartTCPServer()

	log.Println("Starting Peer y")
	y := gotalk.NewManager(true, "localhost:9092")
	y.AddService("echo", echo)
	y.StartTCPServer()

	for i := 0; i < 4; {
		r, err := x.Request("echo", []byte("Hello"))
		log.Println("echo Response", string(r), err)
		time.Sleep(time.Second)
		i++
	}

}

func echo(s *gotalk.Sock, name string, in []byte) ([]byte, error) {
	log.Println("echo received on ", s.Addr(), name, string(in))

	return in, nil
}

func onWSConnect(s *gotalk.Sock) {
	s.Notify("hello", "world")
	log.Println("onWSConnect", s.Addr())
}
