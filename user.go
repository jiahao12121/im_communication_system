package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

//创建一个新建用户的API
func NewUser(conn net.Conn) *User {
	user := &User{
		Name: conn.RemoteAddr().String(),
		Addr: conn.RemoteAddr().String(),
		C:    make(chan string),
		conn: conn,
	}

	//启动监听用户消息通道的goroutine
	go user.ListenMessage()
	return user
}

//监听User中的C通道，一旦有消息，马上发给对应客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		_, _ = user.conn.Write([]byte(msg + "\r\n"))
	}

}
