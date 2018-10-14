package impl

import (
	"github.com/gorilla/websocket"
)

type Connection struct {
	wsConn  *websocket.Conn
	inChan  chan []byte
	outChan chan []byte
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:  wsConn,
		inChan:  make(chan []byte, 1000),
		outChan: make(chan []byte, 1000),
	}

	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {
	data = <-conn.inChan
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	conn.inChan <- data
	return
}
func (conn *Connection) Close() {
	//线程安全的
	conn.wsConn.Close()
}
