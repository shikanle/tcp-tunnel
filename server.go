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

	// the tunnel may be closed or timeout, choose one available
	var tunnel net.Conn
	for {
		tunnel = <-server.connnectionPool
		_, err := tunnel.Read(make([]byte, 0))
		if err == nil {
			break
		}
	}

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
	log.Println("begin refresh")
	for i := len(server.connnectionPool); i > 0; i-- {
		tunnel := <-server.connnectionPool
		_, err := tunnel.Read(make([]byte, 0))
		log.Println("testing", err)
		if err == nil {
			log.Println("Good", tunnel, i)

			tunnel.SetReadDeadline(time.Now().Add(time.Second))
			_, e2 := tunnel.Read(make([]byte, 1))
			log.Println(e2.Error())

			server.connnectionPool <- tunnel
		} else {
			log.Println("Bad", tunnel, i)
		}
	}
	log.Println("end refresh")

	select {
	case server.connnectionPool <- conn:
		log.Printf("[Pool = %d] Connection established", len(server.connnectionPool))
	default:
		conn.Close()
		log.Printf("[Pool = %d] Connection rejected, queue full", len(server.connnectionPool))
	}
}

func (server *TcpTunnelServer) Serve() {
	go server.serve(server.clientPort, server.handleClientConnection)
	go server.serve(server.tunnelPort, server.handleTunnelConnection)
}
