package net

import (
	"encoding/binary"
	"log/slog"
	"net"
	"time"
)

type Client struct {
	conn net.Conn
}

func NewClient(net net.Conn) *Client {
	return &Client{conn: net}
}

func (c *Client) Send(data string) error {
	dataLen := len(data)
	sendData := make([]byte, dataLen+numOfBytes)
	binary.BigEndian.PutUint64(sendData[:numOfBytes], uint64(dataLen))
	copy(sendData[numOfBytes:], data)
	if _, err := c.conn.Write(sendData); err != nil {
		slog.Error("go-micro: Connect Write fail", "error", err)
		return err
	}

	gotDataLen := make([]byte, numOfBytes)
	if _, err := c.conn.Read(gotDataLen); err != nil {
		slog.Error("go-micro: handleConn read dataLen error", "error", err)
		return err
	}
	length := binary.BigEndian.Uint64(gotDataLen)
	gotData := make([]byte, length)
	if _, err := c.conn.Read(gotData); err != nil {
		slog.Error("go-micro: handleConn read data error", "error", err)
		return err
	}

	slog.Info("go-micro: Connect success", "sendData", data, "gotData", string(gotData))
	return nil
}

func Connect(network, address string) error {
	conn, err := net.DialTimeout(network, address, time.Second*10)
	if err != nil {
		slog.Error("go-micro: Connect DialTimeout fail", "error", err)
		return err
	}
	defer func() {
		_ = conn.Close()
	}()
	for {
		sendData := "hello world"
		if _, err := conn.Write([]byte(sendData)); err != nil {
			slog.Error("go-micro: Connect Write fail", "error", err)
			return err
		}
		gotData := make([]byte, 1024)
		if _, err := conn.Read(gotData); err != nil {
			slog.Error("go-micro: Connect Read fail", "error", err)
			return err
		}
		slog.Info("go-micro: Connect success", "sendData", sendData, "gotData", string(gotData))
	}

}
