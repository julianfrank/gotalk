package main

import (
	"github.com/julianfrank/gotalk"
	"log"
)

func main() {
	log.Println("Started")
	x := gotalk.NewManager()
	x.AddHandler("echo", echo)
	log.Println(x)
}

func echo(in []byte) ([]byte, error) {
	return in, nil
}
