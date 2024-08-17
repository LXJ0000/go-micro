package main

import (
	"log/slog"
	"time"

	"github.com/LXJ0000/go-micro/net"
)

func main() {
	go func() {
		if err := net.Serve("tcp", ":8080"); err != nil {
			slog.Error("main: serve fail", "error", err)
		}
	}()
	time.Sleep(time.Second * 10)

	if err := net.Connect("127.0.0.1", ":8080"); err != nil {
		slog.Error("main: connect fail", "error", err)
	}
}
