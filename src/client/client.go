package main

import (
	"bytes"
	"conf"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
)

func main() {
	//clients := getClients()
	//模拟多个物理设备同时发送多条命令
	sendData := getSendData(1,1)
	for clientId, msg := range sendData {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", "127.0.0.1", "8080"))
		if err != nil {
			panic(err.Error())
		}
		for msgId, msg := range msg {
			//包头,3字节,为了减少包头和数据内容重叠概率，采用3字节包头， 为DNY
			var headb = make([]byte, 3)
			headb = []byte(conf.HEADER)
			//物理ID，4字节,设备物理ID和设备上的二维码中的一致，无重复，不丢失
			var clientb = make([]byte, 4)
			if conf.BigEndian {
				binary.BigEndian.PutUint32(clientb, uint32(clientId))
			}else {
				binary.LittleEndian.PutUint32(clientb, uint32(clientId))
			}

			//消息ID，2字节,每条命令的消息ID必须不同，但重发的命令的消息ID必须相同
			var msgIdb = make([]byte, 2)
			if conf.BigEndian {
				binary.BigEndian.PutUint16(clientb, uint16(msgId))
			}else {
				binary.LittleEndian.PutUint16(clientb, uint16(msgId))
			}
			//命令,1字节
			var cmdb = make([]byte, 1)
			cmdb = []byte{0x01}
			//数据,n字节,n<248
			var msgb = make([]byte, len(msg))
			msgb = []byte(msg)
			//校验,2字节，为了保证每条命令传输的正确性，采用2字节无符号累加和校验，校验从包头到数据的内容
			var crcb = make([]byte, 2)
			crcb = []byte{0x12,0x34}

			//长度,2字节,长度=物理ID(4)+消息ID(2)+命令(1)+数据(n) +校验(2)，每包最多256字节
			var lengthb = make([]byte, 2)
			length := len(clientb) + len(msgIdb) + len(cmdb) + len(msgIdb) + len(msg) + len(crcb)
			//将长度值处理成16bit无符号整数，以大端/小端模式放到字节数组
			if conf.BigEndian {
				binary.BigEndian.PutUint16(lengthb, uint16(length))
			}else {
				binary.LittleEndian.PutUint16(lengthb, uint16(length))
			}

			var bufferClient bytes.Buffer
			//包头
			bufferClient.Write(headb)
			//长度,2字节,长度=物理ID+消息ID+命令+数据(n) +校验(2)，每包最多256字节
			bufferClient.Write(lengthb)
			//物理ID
			bufferClient.Write(clientb)
			//消息ID
			bufferClient.Write(msgIdb)
			//命令
			bufferClient.Write(cmdb)
			//数据
			bufferClient.Write(msgb)
			//校验
			bufferClient.Write(crcb)

			b3 := bufferClient.Bytes()
			_, err := conn.Write(b3)
			if err != nil {
				panic(err)
			}

			log.Println("ID为",clientId,"的设备发送第", msgId, "个消息","完整内容",b3)
		}
	}
}
//
//func getClients() []string {
//	return []string{"1","2","3","4"}
//}
func getSendData(clientNum int, cmdNum int)(data [][]string) {
	randomDatas := make([]string, 256)
	for i := 0; i < 256; i++ {
		randomData := ""
		for j := 0; j <= i; j ++ {
			randomData += strconv.Itoa(rand.Intn(16))
		}
		randomDatas[i] = randomData
	}
	data = make([][]string,clientNum)
	for i := 0; i < clientNum; i ++ {
		row := make([]string,cmdNum)
		for j := 0; j < cmdNum; j++ {
			row[j] = fmt.Sprintf("%v-%v:%s",i,j,randomDatas[rand.Intn(256)])
		}
		data[i] = row
	}
	return
}
