package protocol

import (
	"conf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type Buffer struct {
	reader         io.Reader "从reader中读取数据到buf中"
	header         string "通过header来标识一个完整报文的开端"
	headLengthSize int "报文长度标识的长度，从buf中取headLengthSize长度的数据转成int，就可以获得报文的长度"
	buf            []byte "使用buf作为报文缓冲，处理分包粘包"
	bufLength      int "buf的最大长度，当一次读取的数据长度大于该值时"
	start          int "buf中有效数据区域的起点，buf[0,start)为无效数据,buf[start,end]为可读数据,buf[end+1,bufLength)为可写部分"
	end            int "buf中有效数据区域的终点"
}
//初始化一个Buffer
func NewBuffer(reader io.Reader) *Buffer {
	buf := make([]byte, conf.BufferLength)
	return &Buffer{reader, conf.HEADER, conf.HeadLengthSize, buf, conf.BufferLength, 0, 0}
}

// grow 将有用的字节前移，end -= start,start = 0
// TODO 这是是否可以考虑环形缓冲区，避免数据搬移操作
func (buffer *Buffer) grow() {
	if buffer.start == 0 {
		return
	}
	copy(buffer.buf, buffer.buf[buffer.start:buffer.end])
	buffer.end -= buffer.start
	buffer.start = 0
}

//end-start就是可读数据的长度
func (buffer *Buffer) len() int {
	return buffer.end - buffer.start
}

//返回n个字节，而不产生移位
func (buffer *Buffer) seek(n int) ([]byte, error) {
	if buffer.end-buffer.start >= n {
		buf := buffer.buf[buffer.start : buffer.start+n]
		return buf, nil
	}
	return nil, errors.New("not enough")
}

//将start前移offset位，舍弃offset个字段，从新的start开始读取n个字段，然后再将start前移n位
func (buffer *Buffer) read(offset, n int) []byte {
	buffer.start += offset
	buf := buffer.buf[buffer.start : buffer.start+n]
	buffer.start += n
	return buf
}

//从reader里面读取数据，如果当前end已经达到bufLength，则认为读取失败，读取成功后end += n，即buf[start,end)为本次读取的数据
//如果reader阻塞，会发生阻塞
func (buffer *Buffer) readFromReader() error {
	if buffer.end >= buffer.bufLength {
		log.Println("一个完整的数据包太长已经超过你定义的example.BUFFER_LENGTH ", buffer.bufLength)
		return errors.New(fmt.Sprintf("一个完整的数据包太长已经超过你定义的example.BUFFER_LENGTH(%d)\n", buffer.bufLength))
	}
	n, err := buffer.reader.Read(buffer.buf[buffer.end:])
	if err != nil {
		log.Println("读取出错了", err, "但还是读取到了", n, "字节数据")
		return err
	}
	log.Println("读取到了", n, "字节数据")
	buffer.end += n
	return nil
}

//用一个无限循环读取有效数据，读取出错或者消息处理出错时结束循环，每次读取前先把start和end间的数据搬移到[0,end-start]
func (buffer *Buffer) Read(msg chan string) error {
	for {
		log.Println("开始读取数据到buffer")
		buffer.grow()
		err1 := buffer.readFromReader() // 读数据拼接到定额缓存后面
		if err1 != nil {
			return err1
		}
		err2 := buffer.checkMsg(msg)
		if err2 != nil {
			return err2
		}
	}
}

// 检查定额缓存里面的数据有几个消息(可能不到1个，可能连一个消息头都不够，可能有几个完整消息+一个消息的部分)，
// 然后把有效的消息放到msg中
func (buffer *Buffer) checkMsg(msg chan string) error {
	//消息头长度=len("BEGIN") + 4 假设消息头标识为"BEGIN"
	HeaderLeng := len(buffer.header) + buffer.headLengthSize
	//查看消息头长度的内容
	headBuf, err1 := buffer.seek(HeaderLeng)
	if err1 != nil { // 一个消息头都不够， 跳出去继续读吧, 但是这不是一种错误
		return nil
	}
	// 判断消息头正确性，看[0,len("BEGIN"))的内容是否为"BEGIN"
	if string(headBuf[:len(buffer.header)]) == buffer.header {
		log.Println("读取到了消息头")
	} else {
		log.Println("没有读取到消息头")
		return errors.New("消息头部不正确")
	}
	//大端存储
	//从消息头结束位开始，获取16bit，然后转成int，即为消息体长度标志位
	var contentSize int
	if conf.BigEndian {
		contentSize = int(binary.BigEndian.Uint16(headBuf[len(buffer.header):]))
	}else {
		contentSize = int(binary.LittleEndian.Uint16(headBuf[len(buffer.header):]))
	}
	log.Println("读取到的消息长度为:", contentSize)
	//如果缓冲中的可读长度大于消息体长度，表示这里可以读到一个有效的消息体
	if buffer.len() >= contentSize-HeaderLeng {
		// 把消息读出来，把start往后移
		contentBuf := buffer.read(len(buffer.header), contentSize)
		content := []byte(contentBuf)
		log.Println("读取到的消息为:", content)
		msg <- string(content)
		// 递归，看剩下的还够一个消息不
		err3 := buffer.checkMsg(msg)
		if err3 != nil {
			return err3
		}
	} else {
		log.Println("buffer内数据长度", buffer.len(), " 小于消息体长度 ", contentSize)
	}
	return nil
}
