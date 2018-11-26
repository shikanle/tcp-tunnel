package main

import (
	"flag"
	"time"
)

var (
	argMode        = flag.String("mode", "server", "Mode of the service: server or client")
	argPublishPort = flag.Int("publish", 8080, "[Server] publish listening port, e.g., 6000")
	argTunnelPort  = flag.Int("tunnel", 7000, "[Server] tunnel listening port, e.g., 7000")
	argServerUri   = flag.String("server", "localhost:7000", "[Client] server host and port, e.g., localhost:7000")
	argLocalUri    = flag.String("local", "13.232.248.187:18083", "[Client] local host and port, e.g., localhost:80")
	argPoolSize    = flag.Int("pool", globalConnectionPoolSize, "[Server/Client] connection pool size, e.g., 16")
)

func main() {
	flag.Parse()
	if *argMode == "server" {
		s := NewTcpTunnelServer(*argPublishPort, *argTunnelPort, *argPoolSize)
		s.Serve()
	} else {
		s := NewTcpTunnelClient(*argServerUri, *argLocalUri, *argPoolSize)
		s.Serve()
	}
	for {
		time.Sleep(time.Second)
	}
}
