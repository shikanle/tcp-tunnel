package main

import (
	"log"
	"net"
	"strconv"
	"time"
)

type TcpTunnelServer struct {
	clientPort      int
	tunnelPort      int
	connnectionPool chan net.Conn
}

func NewTcpTunnelServer(clientPort int, tunnelPort int, poolSize int) *TcpTunnelServer {
	server := new(TcpTunnelServer)
	server.clientPort = clientPort
	server.tunnelPort = tunnelPort
	server.connnectionPool = make(chan net.Conn, poolSize)
	return server
}

func (server *TcpTunnelServer) serve(listenPort int, handler func(net.Conn)) {
	for {
		listenTo := ":" + strconv.Itoa(listenPort)
		l, err := net.Listen("tcp", listenTo)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Listening to", listenTo, "...")
			for {
				c, err := l.Accept()
				if err != nil {
					log.Println(err)
					break
				}
				go handler(c)
			}
		}
		time.Sleep(time.Second)
		log.Println("Reconnecting...")
	}
}

func (server *TcpTunnelServer) handleClientConnection(conn net.Conn) {
	log.Printf("[Pool = %d] Visiting connection established", len(server.connnectionPool))
	tunnel := <-server.connnectionPool
	go func() {
		buf := make([]byte, globalDefaultBufferSize)
		for {
			read, err := tunnel.Read(buf)
			if err != nil || read <= 0 {
				break
			}
			if read == globalDefaultBufferSize {
				conn.Write(buf)
			} else {
				conn.Write(buf[:read])
			}
		}
		tunnel.Close()
		conn.Close()
	}()

	buf := make([]byte, globalDefaultBufferSize)
	for {
		read, err := conn.Read(buf)
		if err != nil || read <= 0 {
			break
		}
		if read == globalDefaultBufferSize {
			tunnel.Write(buf)
		} else {
			tunnel.Write(buf[:read])
		}
	}
	tunnel.Close()
	conn.Close()
	log.Printf("[Pool = %d] Connection closed", len(server.connnectionPool))
}

func (server *TcpTunnelServer) handleTunnelConnection(conn net.Conn) {
	server.connnectionPool <- conn
	log.Printf("[Pool = %d] Connection established", len(server.connnectionPool))
}

func (server *TcpTunnelServer) Serve() {
	go server.serve(server.clientPort, server.handleClientConnection)
	go server.serve(server.tunnelPort, server.handleTunnelConnection)
}
