package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"os"
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
func main() {                                //main函数
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
func (s *FTPserver)startPassiveListener()(net.Listener,int,error) {              //为被动模式创建一个临时监听器
	datalinstener,err  := net.Listener("tcp","127.0.0.1:0")        // 监听本地任意可用端口
	if err != nil {
		return nil,0,err
	}
	_, portStr, _ := net.SplitHostPort(dataListener.Addr().String())     // 获取实际端口号
	port, _ := strconv.Atoi(portStr)
	return dataListener, port, nil
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
	var dataListener net.Listener


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
			// 启动被动监听
			dataListener, port, err := s.startPassiveListener()
			if err != nil {
				conn.Write([]byte("425 Can't open data connection.\r\n"))
				continue
			}
			p1 := port / 256
			p2 := port % 256              // 转成 127,0,0,1,p1,p2       FTP协议要求是p1*256+p2=端口号

			response := fmt.Sprintf("227 Entering Passive Mode (127,0,0,1,%d,%d)\r\n", p1, p2)

			conn.Write([]byte(response))

			// 把 dataListener 存起来，以便 LIST 时使用

		}


		if line == "LIST" {                       //实现 LIST 命令
			if dataListener == nil {
				conn.Write([]byte("425 Use PASV first.\r\n"))
				continue
			}


			conn.Write([]byte("150 Opening data connection.\r\n"))

			// 等客户端连上数据端口
			dataConn, err := dataListener.Accept()
			if err != nil {
				conn.Write([]byte("425 Can't open data connection.\r\n"))
				// 出错就清理掉数据监听器
				dataListener.Close()
				dataListener = nil
				continue
			}

			// 读取当前目录的文件和文件夹名字
			files, err := os.ReadDir(".")
			if err != nil {
				conn.Write([]byte("550 Failed to read directory.\r\n"))  //550 请求操作未执行：文件不可用
				dataConn.Close()
				dataListener.Close()
				dataListener = nil
				continue
			}

			// 把文件名一个一个发过去（最简单的格式：每个名字一行）
			for _, file := range files {
				name := file.Name()
				if file.IsDir() {
					name += "/"   // 文件夹后面加个斜杠，方便辨认
				}
				dataConn.Write([]byte(name + "\r\n"))
			}

			// 关数据连接，关数据监听器
			dataConn.Close()
			dataListener.Close()
			dataListener = nil
			conn.Write([]byte("226 Transfer complete.\r\n"))
			continue
		}

		if strings.HasPrefix(line, "CWD ") {     //实现 CWD（切换目录）
			dir := strings.TrimSpace(strings.TrimPrefix(line, "CWD "))
			err := os.Chdir(dir)
			if err != nil {
				conn.Write([]byte("550 Failed to change directory.\r\n"))   //切换目录失败 ，然后 550 FTp命令意思是   550 请求操作未执行：文件不可用

			} else {
				conn.Write([]byte("250 Directory changed.\r\n"))   //目录切换  250 意思是 250 请求文件操作成功
			}
			continue
		}

		if line == "CDUP" {           //返回上一级目录
			err := os.Chdir("..")      //对应操作系统里面的 cd ..命令
			if err != nil {
				conn.Write([]byte("550 Failed to change directory.\r\n"))   //切换目录失败 ，然后 550 FTp命令意思是   550 请求操作未执行：文件不可用
			} else {
				conn.Write([]byte("250 Directory changed.\r\n"))     //250 请求文件操作成功
			}
			continue
		}

		if strings.HasPrefix(line,"RETR") {     //客户端发 RETR hello.txt，服务器就把 hello.txt 的内容通过数据连接发给客户端。
			filename := strings.TrimSpace(strings.TrimPrefix(line,"RETR"))
			if dataListener == nil{
				conn.Write([]byte("425 Use PASV first.\r\n"))   //425 无法打开数据连接
				continue
			}

			file,err := os.Open(filename)   //把“文件名字符串”翻译成操作系统能操作的文件对象，然后才可以从里面读数据
			if err != nil {
				conn.Write([]byte("550 File not found.\r\n"))    //550 请求操作未执行：文件不可用
				continue
			}
			defer file.Close()

			conn.Write([]byte("150 Opening data connection.\r\n"))   //150 文件状态正常，即将打开数据连接
			dataconn,err := dataListener.Accept()
			if err != nil{
				conn.Write([]byte("425 Can't open data connection.\r\n"))  //425 无法打开数据连接
				dataListener.Close()
				dataListener = nil         //这里强制把datalistener设置为nil，是为了刷新状态，使客户端重新走一步PASV流程
				continue
			}
			io.Copy(dataconn,file)  //把文件内容丢给数据连接

			dataconn.Close()
			dataListener.Close()   //关闭 ≠ 忘记。后续还需要重置状态 对应上面的if dataListener != nil
			dataListener = nil
			conn.Write([]byte("226 Transfer complete.\r\n"))  //   226 关闭数据连接，请求操作成功
			continue
		}

		// ------------------------------------------实现文件的上传

		if strings.HasPrefix(line, "STOR ") {
			filename := strings.TrimSpace(strings.TrimPrefix(line, "STOR "))
			if dataListener == nil {
				conn.Write([]byte("425 Use PASV first.\r\n"))    //   425 无法打开数据连接
				continue
			}


			file, err := os.Create(filename) // 创建文件
			if err != nil {
				conn.Write([]byte("550 Failed to create file.\r\n"))   //   550 请求操作未执行：文件不可用
				continue
			}
			defer file.Close()

			conn.Write([]byte("150 Opening data connection.\r\n"))  //   150 文件状态正常，即将打开数据连接

			dataconn2, err := dataListener.Accept()
			if err != nil {
				conn.Write([]byte("425 Can't open data connection.\r\n"))   //   425 无法打开数据连接

				dataListener.Close()
				dataListener = nil
				continue
			}

			// 从数据连接读取客户端发来的数据，然后写入文件
			io.Copy(file, dataconn2)

			dataconn2.Close()
			dataListener.Close()
			dataListener = nil
			conn.Write([]byte("226 Transfer complete.\r\n"))  //   226 关闭数据连接，请求操作成功
			continue
		}


		//------------------------------创建目录
		if strings.HasPrefix(line, "MKD ") {
			dirname := strings.TrimSpace(strings.TrimPrefix(line, "MKD "))
			err := os.Mkdir(dirname, 0755)                                     //给 os.Mkdir 的权限参数，表示“所有者读写执行，同组和其他人读执行”。在 Windows 上影响很小，但在 Linux 上能让文件夹正常使用。
			if err != nil {
				conn.Write([]byte("550 Failed to create directory.\r\n"))   //   550 请求操作未执行：文件不可用
			} else {
				conn.Write([]byte(fmt.Sprintf(`257 "%s" created.\r\n`, dirname)))   //   257 创建路径名
			}
			continue
		}

		//------------------------------删除空目录
		if strings.HasPrefix(line, "RMD ") {
			dirname := strings.TrimSpace(strings.TrimPrefix(line, "RMD "))
			err := os.Remove(dirname)      //用 os.Remove 删除（它只能删除空目录，如果目录里有文件会报错）
			if err != nil {
				conn.Write([]byte("550 Failed to remove directory.\r\n"))    //   550 请求操作未执行：文件不可用
			} else {
				conn.Write([]byte("250 Directory removed.\r\n"))         //   250 请求文件操作成功
			}
			continue
		}
		//------------------------------删除文件
		if strings.HasPrefix(line, "DELE ") {
			filename := strings.TrimSpace(strings.TrimPrefix(line, "DELE "))
			err := os.Remove(filename)   //删除文件就没问题
			if err != nil {
				conn.Write([]byte("550 Failed to delete file.\r\n"))   //   550 请求操作未执行：文件不可用
			} else {
				conn.Write([]byte("250 File deleted.\r\n"))     //   250 请求文件操作成功
			}
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
