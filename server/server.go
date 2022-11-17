package main

import (
	"encoding/binary"
	"fmt"
	"jnet/network/base"
	"jnet/network/tcp"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func test(request base.IRequest) bool {
	switch request.GetMsgID() {
	case 1:
		fmt.Println("connect ", request.GetConnection().ID())
	case 2:
		fmt.Println("close ", request.GetConnection().ID())
	}
	return true
}

func main() {
	tcpServer := tcp.NewServer("127.0.0.1:1440")
	tcpServer.BindPacketFunc(test)
	tcpServer.Serve()
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	time.Sleep(5 * time.Second)
	ClientTest(1)
	s := <-c
	fmt.Printf("server exit ------- signal:[%v]", s)
}

func ClientTest(i uint32) {
	//conn, err := kcp2.DialWithOptions("127.0.0.1:8765", nil, 0, 0)
	//if err != nil {
	//	fmt.Println("client start err, exit! ", err)
	//	return
	//}
	rand.Seed(time.Now().Unix())
	var conns []net.Conn
	for i := 0; i < 1000; i++ {
		conn, err := net.Dial("tcp4", "127.0.0.1:1440")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(10 * time.Second)
		}
		conns = append(conns, conn)

	}
	dp := &base.PacketParser{
		PacketHeadLen: 8, //uint32+uint32
		MaxPacketLen:  40960,
		ByteOrder:     binary.BigEndian,
	}
	var count = 300
	//for i := 0; i < count; i++ {
	//	go func(i uint32) {
	//		for _, v := range conns {
	//			length := rand.Intn(1000)
	//			msg, _ := dp.Encode(base.NewMsgPackage(uint32(i), []byte(RandStringBytes(length))))
	//			_, err := v.Write(msg)
	//			if err != nil {
	//				fmt.Println("client write err: ", err)
	//				return
	//			}
	//			time.Sleep(10 * time.Millisecond)
	//
	//		}
	//	}(uint32(i))
	//}
	for _, v := range conns {
		go func(v net.Conn) {
			for i := 0; i < count; i++ {
				length := rand.Intn(1000)
				msg, _ := dp.Encode(base.NewMsgPackage(uint32(i), []byte(RandStringBytes(length))))
				_, err := v.Write(msg)
				if err != nil {
					//fmt.Println("client write err: ", err)
					return
				}

				time.Sleep(10 * time.Second)
			}
		}(v)
	}
}

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = StdChars[rand.Intn(len(StdChars))]
	}
	return string(b)
}
