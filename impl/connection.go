package impl

import (
	//"time"
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
)

var connMap map[string]*Connection = make(map[string]*Connection, 1000)

type Connection struct {
	wsConnection *websocket.Conn //websocket链接的
	inChan       chan []byte     //读取的数据
	outChan      chan []byte     //需要写入的数据
	closeChan    chan byte

	mutex    sync.Mutex
	isClosed bool

	uid string
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConnection: wsConn,
		inChan:       make(chan []byte, 1000),
		outChan:      make(chan []byte, 1000),
		closeChan:    make(chan byte, 1),
		isClosed:     false,
		uid:          "",
	}

	var uid string
	if uid, err = conn.GetUid(); err != nil {
		conn.Close()
		return
	}

	logs.Debug("getuid:" + uid)

	connMap[uid] = conn

	go conn.readLoop()
	go conn.writeLoop()

	return
}

/**
*生成链接的标识符
 */
func (conn *Connection) GetUid() (uid string, err error) {
	if len(conn.uid) == 0 {
		conn.uid = conn.wsConnection.RemoteAddr().String()
	}
	uid = conn.uid

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
	case conn.outChan <- data:
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

		if err = conn.wsConnection.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) Broadcast(data []byte) (err error) {
	/**
	广播消息的
	*/
	//var conn1 *Connection
	for _, conn1 := range connMap {

		tmpuid, _ := conn1.GetUid()
		logs.Debug("broadcast message:" + tmpuid)

		if err = conn1.WriteMessage(data); err != nil {
			conn1.Close()

			return
		}

	}

	return
}

func (conn *Connection) Close() {
	//线程安全的
	conn.wsConnection.Close()

	//加锁保证线程安全的
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()

	//删除全局的链接
	uid, _ := conn.GetUid()
	if _, ok := connMap[uid]; ok {
		delete(connMap, uid)
	}
}
