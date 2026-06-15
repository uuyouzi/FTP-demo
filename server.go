package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Config struct {
	ip       string
	port     string
	username string
	password string
}
type FTPserver struct {
	cfg      *Config
	listener net.Listener
}

func Addconfig(cfg *Config) {
	cfg.ip = "127.0.0.1"
	cfg.port = "21"
	cfg.username = "youzi"
	cfg.password = "uuyouzi"
}
func main() {
	cfg := &Config{}
	Addconfig(cfg)

	address := net.JoinHostPort(cfg.ip, cfg.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	s := &FTPserver{
		cfg:      cfg,
		listener: listener,
	}
	s.Server()
}
func (s *FTPserver) Server() error {
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
	conn.Write([]byte("220 Service ready\r\n"))

	//data, err := reader.Peek(16)
	//fmt.Println(err)
	//fmt.Println(data)
	//fmt.Println(string(data))
	var login bool
	var username string


	for {
		// 3. 读取客户端发来的一行命令（以 \r\n 结束）
		line, err := reader.ReadString('\n')
		if err != nil {
			// 客户端断开或读取出错，退出循环
			break
		}
		line = strings.TrimSpace(line) //把字符串前后所有看不见的空白符号（空格、\r、\n、tab）统统擦掉
		fmt.Println("收到命令:", line)

		if strings.HasPrefix(line,"USER"){
			username = strings.TrimSpace(strings.HasPrefix(line,"USER"))
			conn.Write([]byte("331 输入密码 \r\n"))
			continue
		}

		if strings.HasPrefix(line."PASS"){
			password = strings.TrimSpace(strings.HasPrefix(line,"PASS"))

			if username == s.cfg.username && password == s.cfg.password{
				login = true
				conn.Write([]byte("230 登录成功。\r\n"))
				continue
			}else {
				conn.Write([]byte("530 登录失败。\r\n"))
				return
			}
		}
		if !login {
			conn.Write([]byte("530 请登录账号。 \r\n"))
			continue
		}

		if line == "QUIT"{
			conn.Write([]byte("221 Goodbye\r\n"))
			break
		}else {
			conn.Write([]byte("502 命令未实现。\r\n"))
		}





		//// 4. 处理命令（目前只是简单回显）
		//if line == "QUIT" {
		//	conn.Write([]byte("221 Goodbye\r\n"))
		//	break
		//} else {
		//	// 其他命令返回 502 未实现
		//	conn.Write([]byte("502 Command not implemented\r\n"))
		//}
	}

}
