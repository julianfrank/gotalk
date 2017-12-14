package main

import (
	"log"

	"github.com/julianfrank/gotalk"
)

func main() {
	//startTime := time.Now()
	log.Println("Starting Peer x")
	x := gotalk.NewManager(true, "localhost:9090")
	x.AddService("echo", echo)
	x.StartTCPServer()

	log.Println("Starting Peer y")
	y := gotalk.NewManager(true, "localhost:9092")
	y.AddService("echo", echo)
	y.StartTCPServer()

	r, err := x.Request("echo", []byte("Hello"))
	log.Println("echo Response", string(r), err)

}

func echo(s *gotalk.Sock, name string, in []byte) ([]byte, error) {
	log.Println("echo received on ", s.Addr(), name, string(in))

	return in, nil
}

func onWSConnect(s *gotalk.Sock) {
	s.Notify("hello", "world")
	log.Println("onWSConnect", s.Addr())
}
