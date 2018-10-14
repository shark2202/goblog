package main

import (
	"net/http"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"github.com/zd04/goblog/impl"
)

var (
	upgrader = websocket.Upgrader{
		//设置允许跨域的
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {

	var (
		wsConn *websocket.Conn
		conn   *impl.Connection

		err  error
		data []byte
		//data interface{}
	)

	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	logs.Debug("Upgrade")

	//保存链接的
	if conn, err = impl.InitConnection(wsConn); err != nil {
		goto ERR
	}

	for {
		//if err = conn.ReadJSON(&data); err != nil {
		//	goto ERR
		//}
		if _, data, err = wsConn.ReadMessage(); err != nil {
			goto ERR
		}

		logs.Debug(string(data))

		if err = wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR:
	logs.Debug(err)
	logs.Debug("close")
	conn.Close()
	//wsConn.Close()
}

func startWsServer() error {

	//
	logs.Debug("StartWsServer")

	go func() {
		logs.Debug("on :7777")
		http.HandleFunc("/ws", wsHandler)
		http.ListenAndServe("0.0.0.0:7777", nil)
	}()

	return nil
}
