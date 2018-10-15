package impl

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	wsConn    *websocket.Conn //websocket链接的
	inChan    chan []byte     //读取的数据
	outChan   chan []byte     //需要写入的数据
	closeChan chan byte

	mutex    sync.Mutex
	isClosed bool
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:  wsConn,
		inChan:  make(chan []byte, 1000),
		outChan: make(chan []byte, 1000),
	}

	go conn.readLoop()
	go conn.writeLoop()

	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {

	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("conn is closed")
	}

	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {

	select {
	case conn.inChan <- data:
	case <-conn.closeChan:
		err = errors.New("conn is closed")
	}
	return
}

func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)

	for {
		if _, data, err = conn.wsConnection.ReadMessage(); err != nil {
			goto ERR
		}

		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}

		if err = conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) Close() {
	//线程安全的
	conn.wsConn.Close()

	//加锁保证线程安全的
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.nutex.Unlock()
}
