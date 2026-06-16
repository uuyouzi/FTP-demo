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

		if strings.HasPrefix(line,"USER "){             //USER它是一个固定的前缀,客户端发过来的前缀,后面有个空格区分命令和参数 ，客户端发过来会是 USER youzi这样

			username = strings.TrimSpace(strings.TrimPrefix(line,"USER "))

			conn.Write([]byte("331 输入密码 \r\n"))
			continue
		}

		if strings.HasPrefix(line,"PASS "){            //PASS  一个固定的前缀, 客户端发过来的前缀 这几行代码为了提取和验证账号密码用，但是有个坑windows客户端会先用匿名登录（发 USER anonymous），如果失败了才会弹窗让输入用户名密码。

			password := strings.TrimSpace(strings.TrimPrefix(line,"PASS "))

			if username == s.cfg.username && password == s.cfg.password{

				login = true

				conn.Write([]byte("230 登录成功。\r\n"))
				continue
			}else {
				conn.Write([]byte("530 登录失败。\r\n"))
				return                           //没通过验证，直接切断联系所以用return
			}
		}

		if !login {
			conn.Write([]byte("530 请登录账号。 \r\n"))          //登录状态
			continue
		}

		if line == "PWD" {
			conn.Write([]byte("257 \"/\" 是当前目录。\r\n"))              //257 对应FTP响应码 意思是创建路径名
			continue
		}

		if line == "OPTS utf8 on"{                              //选项命令：启用 UTF‑8
			conn.Write([]byte("200 成功\r\n"))
			continue
		}
		if line == "SYST" {
			conn.Write([]byte("215 UNIX Type: L8\r\n"))             //响应 "215 UNIX Type: L8"  意思是系统类型  可译为 “215 UNIX 类型：L8”，其中 L8 表示逻辑字节大小为 8 位（标准二进制流）
			continue
		}
		if line == "SITE help" {
			conn.Write([]byte("214 SITE help\r\n"))               //214对应帮助信息
			continue
		}
		if line == "TYPE A" || line == "TYPE I" {              //客户端询问传输文本模式或者二进制模式
			conn.Write([]byte("200 Type ok\r\n"))
			continue
		}
		if line == "NOOP" {
			conn.Write([]byte("200 NOOP ok.\r\n"))          //NOOP是心跳检测，检查服务器是否还活着这里返回一个200的命令
			continue
		}
		if line == "PASV" {
			conn.Write([]byte("227 进入被动模式 (127,0,0,1,4,20)\r\n"))     // 稍后实现真正的数据端口，现在先回个占位响应，防止客户端卡死
			continue
		}


		if line == "QUIT" {
			conn.Write([]byte("221 Goodbye\r\n"))           //结束连接  221表示 服务关闭控制连接
			break
		}else {
			conn.Write([]byte("502 命令未实现。\r\n"))          // 其他命令返回 502 未实现
		}






	}

}
