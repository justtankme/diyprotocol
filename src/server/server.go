package main

import (
	"conf"
	"fmt"
	"log"
	"net"
	"os"
	"protocol"
	"sync"
)

func main() {
	addr := fmt.Sprintf("%s:%s", "127.0.0.1", "8080")
	listen, err1 := net.Listen("tcp", addr)
	if err1 != nil {
		log.Println("Error: ", err1.Error())
		os.Exit(1)
	}
	for {
		conn, err2 := listen.Accept() // 主goroutine阻塞在这里
		if err2 != nil {
			log.Println("Error: ", err2.Error())
			os.Exit(1)
		}
		go handle(conn) // 开辟新的routine处理每一个客户端
	}
}

func handle(conn net.Conn) {
	msg := make(chan string, conf.ChanMsgCount) // 这里设置消息channel可以容纳10个消息
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	buffer1 := protocol.NewBuffer(conn)

	var wg sync.WaitGroup
	wg.Add(2) // 主的routine将等待两个routine(读消息, 处理消息)的完成
	go func() {
		err := buffer1.Read(msg)
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("收到客户端EOF")
				close(msg) // 对等方关闭了, 这里关闭chan, 通知接收消息的routine别等了，人家都关了
			} else {
				panic(err)
			}
		}
		defer wg.Add(-1)
	}()
	go func() {
		//这里可以对消息内容进行处理
		doMsg(msg)
		defer wg.Add(-1)
	}()
	wg.Wait()
	log.Println("一个客户端处理的消息处理完毕")
}

func doMsg(msg chan string) {
	count := 0
	for v := range msg {
		log.Println("第", count, "个消息体长:", len(v))
		count++
	}
}
