package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	Msg       chan string
	mapLock   sync.RWMutex
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Msg:       make(chan string),
	}
	return server
}

//广播上线消息，即把消息发送到服务器端维护的通道中
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Msg <- sendMsg
}

//必须要有专门监听s.Msg通道的goroutine
//一个无缓存通道的读写必须在两个goroutine中进行
func (s *Server) ListenMsg() {
	for {
		msg := <-s.Msg
		//加锁
		s.mapLock.Lock()
		//把Msg通道中的消息都发到每个User的C通道中
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn)
	//用户上线，先加入map中（map的读写操作需要加锁）
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	//广播上线消息
	s.BroadCast(user, "online!")

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				s.BroadCast(user, "offline!")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}

			//提取用户信息
			msg := string(buf[:n-1])

			//广播用户消息
			s.BroadCast(user, msg)
		}
	}()

	//阻塞当前handler
	select {}
}

func (s *Server) Start() {
	//创建套接字，绑定IP和端口，监听端口三步通过一个函数完成了
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen error")
		return
	}

	//关闭套接字
	defer listener.Close()

	//开启监听s.Msg通道的goroutine
	go s.ListenMsg()
	//接收请求
	for {
		//开始接收客户端请求
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error")
			continue
		}
		//do handler
		//监听到以后就获取了连接，执行具体操作
		go s.Handler(conn)
	}

}
