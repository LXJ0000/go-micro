package net

import (
	"net"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		l, err := net.Listen("tcp", ":8080")
		if err != nil {
			t.Log(err)
			return
		}
		s := NewServer(l)
		if err := s.Start(); err != nil {
			t.Log(err)
			return
		}
		defer func() {
			_ = l.Close()
		}()
	}()
	time.Sleep(time.Second * 3)

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Log(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	c := NewClient(conn)
	if err := c.Send("hello world"); err != nil {
		t.Log(err)
		return
	}

}
