package main

import (
	"bufio"
	"fmt"
	"net"
)

type config struct {
	ip       string
	port     string
	username string
	password string
}
type FTPserver struct {
	cfg      *config
	listener net.Listener
}

func addconfig(cfg *config) {
	cfg.ip = "127.0.0.1"
	cfg.port = "80"
	cfg.username = "佑子"
	cfg.password = "uuuuuuuyouzi"
}
func main() {
	cfg := &config{}
	addconfig(cfg)
	fmt.Println(cfg.ip, cfg.port)
	address := net.JoinHostPort(cfg.ip, cfg.port)
	fmt.Println(address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	s := &FTPserver{
		cfg:      cfg,
		listener: listener,
	}
	s.server()

}
func (s *FTPserver) server() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handerconn(conn)
	}
}
func (s *FTPserver) handerconn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

}
