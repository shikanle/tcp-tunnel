package main

import (
	"log"
	"net"
	"time"
)

type TcpTunnelClient struct {
	localUri           string
	serverUri          string
	waitingConnections chan int
}

func NewTcpTunnelClient(serverUri string, localUri string, poolSize int) *TcpTunnelClient {
	client := new(TcpTunnelClient)
	client.serverUri = serverUri
	client.localUri = localUri
	client.waitingConnections = make(chan int, poolSize)
	return client
}

func (clietn *TcpTunnelClient) visit(local net.Conn, tunnel net.Conn) {
	buffer := make([]byte, globalDefaultBufferSize)
	for {
		read, e := local.Read(buffer)
		if e != nil || read <= 0 {
			break
		} else {
			if read == globalDefaultBufferSize {
				tunnel.Write(buffer)
			} else {
				tunnel.Write(buffer[:read])
			}
		}
	}
	tunnel.Close()
	local.Close()
}

func (client *TcpTunnelClient) tunnelConnectionHandler(conn net.Conn) {
	var local net.Conn = nil
	buffer := make([]byte, globalDefaultBufferSize)
	for {
		read, e := conn.Read(buffer)
		if e != nil || read <= 0 {
			break
		}
		if local == nil {
			log.Println("Visiting", client.localUri)
			local, e = net.Dial("tcp", client.localUri)
			if e != nil {
				break
			}
			go client.visit(local, conn)
		}
		if read == globalDefaultBufferSize {
			local.Write(buffer)
		} else {
			local.Write(buffer[:read])
		}
	}
	if local != nil {
		local.Close()
	}
	conn.Close()
	log.Println("Visiting complete")

	<-client.waitingConnections
	log.Printf("[Pool = %d] Connection closed", len(client.waitingConnections))
}

func (client *TcpTunnelClient) Serve() {
	var connected = true
	var conn net.Conn = nil
	var e error = nil
	for {
		client.waitingConnections <- 0
		for {
			conn, e = net.Dial("tcp", client.serverUri)
			if e == nil {
				break
			}
			if connected {
				log.Println("Failed to connect to", client.serverUri)
				log.Println(e)
				connected = false
			}
			time.Sleep(time.Second)
		}
		tcpconn, ok := conn.(*net.TCPConn)
		if ok {
			tcpconn.SetKeepAlive(true)
			tcpconn.SetKeepAlivePeriod(globalKeepAliveSeconds * time.Second)
		}

		if !connected {
			log.Println("Reconnected to", client.serverUri)
			connected = true
		}
		log.Printf("[Pool = %d] Connection established", len(client.waitingConnections))
		go client.tunnelConnectionHandler(conn)
	}
}
