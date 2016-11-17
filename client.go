package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline          = []byte{'\n'}
	space            = []byte{' '}
	connectedClients map[string]*Client
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
	id   string
	send chan []byte
}

type User struct {
	id      string
	balance int
}

type MessageData struct {
	Typeofdata string                 `json:"type"`
	Data       map[string]interface{} `json:"data"`
}

type ResponseData struct {
	Typeofdata string      `json:"type"`
	Data       interface{} `json:"data"`
}

func (c *Client) readPump() {
	//	defer func() {
	//		c.conn.Close()
	//	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var m MessageData
		//err := c.conn.ReadJSON(&m)
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println("err ", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		json.Unmarshal(message, &m)
		fmt.Printf("%+v\n", m)
		switch m.Typeofdata {
		case "register":
			{
				log.Println("-----")
				col := database.C("user")
				count, er := col.Find(bson.M{"id": c.id}).Count()
				if er != nil {
					log.Println("Count err", er)
				}
				if count == 0 {
					er = col.Insert(&User{c.id, 10000})
					if er == nil {
						fmt.Println("User registered")

					}
				}
				fmt.Println("user", count)

			}
		case "api":
			{
				log.Println("m.Data", m.Data["name"])
				i, e := getQuote(m.Data["name"].(string))
				if e != nil {
					fmt.Println("Error ", e)
				}
				log.Println("-----", i)
				c.conn.WriteJSON(ResponseData{Typeofdata: "api", Data: i})
			}
		}

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) register(id string) {
	isClient := connectedClients[id]
	if isClient != nil {
		log.Println("Client found", id)
	} else {
		c.id = id
		connectedClients[id] = c
		log.Println("New Client Connected", id)
		c.conn.WriteJSON(ResponseData{Typeofdata: "register", Data: "dsf"})
	}

}

func serveWs(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	token := r.URL.Query().Get("token")
	fmt.Println("token", token)

	client := &Client{conn: conn, send: make(chan []byte, 256), id: ""}
	client.register(token)
	go client.writePump()
	client.readPump()

}
