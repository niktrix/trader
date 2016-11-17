package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

//database constant
const (
	dbUserTable  = "user"
	dbStockTable = "stock"

	dbUserId = "userid"
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
	UserId  string
	Balance int
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
				col := database.C(dbUserTable)
				u, er := getUser(c.id)
				if er != nil {
					//create new user
					u.UserId = c.id
					u.Balance = 10000
					er = col.Insert(u)
					if er == nil {
						fmt.Println("User registered")

					}
				}

				c.conn.WriteJSON(ResponseData{Typeofdata: "me", Data: u})

			}
		case "api":
			{
				log.Println("m.Data", m.Data["name"])
				name, ie := m.Data["name"].(string)
				if !ie {
					fmt.Println("Error ", ie)
					break

				}
				i, e := getQuote(name)
				if e != nil {
					fmt.Println("Error ", e)
					break
				}
				if i == nil {
					c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Wrong stock symbol"})
					break
				}
				log.Println("-----", i)
				c.conn.WriteJSON(ResponseData{Typeofdata: "api", Data: i})
			}
		case "buy":
			{
				symbol, b := m.Data["symbol"].(string)
				if !b {
					fmt.Println("Error  mbol].(string)", b)

					break
				}
				quantity, ok := m.Data["quantity"].(float64)
				if !ok {
					fmt.Println("Error  .(float64)", ok)
					break
				}
				log.Println(symbol)
				log.Println(quantity)
				price, company, askerror := getAskPriceAndName(symbol)
				if askerror != nil {
					fmt.Println("askerror ", askerror)
					break
				}
				total := price * float64(quantity)
				fmt.Println(total)

				u, err := getUser(c.id)
				if err != nil {
					fmt.Println("get user err", err)

					break
				}
				fmt.Println(u)

				if float64(u.Balance) >= total {
					fmt.Println("Buy stock")
					u.Balance = u.Balance - int(total)
					err := addStock(&Stock{Quantity: int(quantity), Symbol: symbol, UserId: c.id, PricePaid: total, Company: company})
					if err != nil {
						fmt.Println(err)
					}
					updateUser(&u)
					c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Stock Added"})

				} else {
					c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Not Enough Balance"})

				}

			}
		case "list":
			{
				stocks := getPortfolio(c.id)
				c.conn.WriteJSON(ResponseData{Typeofdata: "list", Data: stocks})

			}

		case "sell":
			{
				symbol, b := m.Data["symbol"].(string)
				if !b {
					fmt.Println("Error  mbol].(string)", b)

					break
				}
				quantity, ok := m.Data["quantity"].(float64)
				if !ok {
					fmt.Println("Error  .(float64)", ok)
					break
				}

				//check if quantity available for sell
				stock, err := getStock(&Stock{Symbol: symbol, UserId: c.id})
				if err != nil {
					c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Stock Not available"})
					break
				}
				if stock.Quantity < int(quantity) {
					c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Stock Added"})
					break

				}

				price, company, askerror := getSellPriceAndName(symbol)
				if askerror != nil {
					fmt.Println("askerror ", askerror)
					break
				}
				total := price * float64(quantity)

				u, err := getUser(c.id)
				if err != nil {
					fmt.Println("get user err", err)

					break
				}
				fmt.Println(u)

				fmt.Println("Sell stock")
				u.Balance = u.Balance + int(total)
				err = removeStock(&Stock{Quantity: int(quantity), Symbol: symbol, UserId: c.id, Company: company})
				if err != nil {
					fmt.Println(err)
				}
				updateUser(&u)
				c.conn.WriteJSON(ResponseData{Typeofdata: "message", Data: "Stock Removed"})

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
		log.Println("Client found", isClient.id)
		c.id = isClient.id
		delete(connectedClients, isClient.id)
		connectedClients[id] = c
	} else {
		c.id = id
		connectedClients[id] = c
		log.Println("New Client Connected", id)
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
