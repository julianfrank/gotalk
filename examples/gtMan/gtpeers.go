package main

import (
	"log"

	"github.com/julianfrank/gotalk"
)

func main() {
	//startTime := time.Now()
	log.Println("Starting Peer x")
	x := gotalk.NewManager(true)
	x.AddHandler("echo", echo)
	x.StartTCPServer("localhost:9090")

	log.Println("Starting Peer y")
	y := gotalk.NewManager(true)
	y.AddHandler("echo", echo)
	y.StartTCPServer("localhost:9092")

	log.Println(" x connecting to y")
	z, err := x.PeerConnect("localhost:9092")

	log.Println(z.Addr(), err)

	r, err := x.Request("echo", []byte("Hello"))
	log.Println(r, err)

	//x.StartWSServer("localhost:9091", onWSConnect, true, nil)

}

func echo(s *gotalk.Sock, name string, in []byte) ([]byte, error) {
	log.Println(s, name, in)
	return in, nil
}

func onWSConnect(s *gotalk.Sock) {
	s.Notify("hello", "world")
	log.Println(s.Addr())
}
