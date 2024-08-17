package net

import (
	"encoding/binary"
	"log/slog"
	"net"
)

const numOfBytes = 8

type Server struct {
	listener net.Listener
}

func NewServer(listener net.Listener) *Server {
	return &Server{listener: listener}
}

func (s *Server) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("go-micro: net: failed to accept connection", "err", err)
			return err
		}
		go func() {
			if err := s.handleConn(conn); err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		// 1. 读取数据长度
		dataLen := make([]byte, numOfBytes)
		if _, err := conn.Read(dataLen); err != nil {
			slog.Error("go-micro: handleConn read dataLen error", "error", err)
			return err
		}
		length := binary.BigEndian.Uint64(dataLen)
		// 2. 读取数据
		data := make([]byte, length)
		if _, err := conn.Read(data); err != nil {
			slog.Error("go-micro: handleConn read data error", "error", err)
			return err
		}
		// 3. 处理数据
		response := s.handleMsg(data)
		// 4. 发送数据
		respLen := len(response)
		respData := make([]byte, numOfBytes+respLen)
		// 4.1 构造发送数据
		binary.BigEndian.PutUint64(respData[:numOfBytes], uint64(respLen))
		copy(respData[numOfBytes:], response)
		// 4.2 发送数据
		if _, err := conn.Write(respData); err != nil {
			slog.Error("go-micro: handleConn write data error", "error", err)
			return err
		}

		slog.Info("go-micro: handleConn write data success")
	}
}

func (s *Server) handleMsg(request []byte) []byte {
	return []byte("hello my friend: " + string(request))
}

func Serve(network, address string) error {
	l, err := net.Listen(network, address)
	if err != nil {
		slog.Error("go-micro: net: failed to listen", "err", err)
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("go-micro: net: failed to accept connection", "err", err)
			return err
		}
		go func() {
			if err := HandleConn(conn); err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func HandleConn(conn net.Conn) error {
	for {
		// 1. 读取数据
		data := make([]byte, 1024)
		if _, err := conn.Read(data); err != nil {
			//// TODO 错误处理...
			//if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			//	slog.Error("go-micro: handleConn closed")
			//	return nil
			//}
			//slog.Warn("go-micro: handleConn read data error", "error", err)
			//continue
			slog.Error("go-micro: handleConn read data error", "error", err)
			return err
		}
		// 2. 处理数据
		response := handleMsg(data)
		// 3. 发送数据
		if _, err := conn.Write(response); err != nil {
			slog.Error("go-micro: handleConn write data error", "error", err)
			return err
		}
		slog.Info("go-micro: handleConn write data success")
	}
}

func handleMsg(request []byte) []byte {
	return request
}
