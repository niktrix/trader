package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
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
	db   *mgo.Database
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

				col := c.db.C(dbUserTable)
				u, er := getUser(c.id, c.db)
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
				log.Println("api called")
				log.Println("m.Data", m.Data["name"])
				name, ie := m.Data["name"].(string)
				if !ie {
					log.Println("Error ", ie)
					break

				}
				i, e := getQuote(name)
				if e != nil {
					log.Println("Error ", e)
					break
				}
				if i == nil {
					log.Println("i == nil ", i)
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
				c.buy(symbol, quantity)

			}
		case "list":
			{
				stocks, err := getPortfolio(c.id, c.db)
				if err != nil {
					log.Println("List", err)
				}
				c.sendResponse(ResponseData{Typeofdata: "list", Data: stocks})

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

				c.sell(symbol, quantity)

			}
		}

	}
}

func (c *Client) buy(symbol string, quantity float64) {
	var (
		err     error
		price   float64
		company string
		user    User
	)
	price, company, err = getAskPriceAndName(symbol)
	if err != nil {
		log.Println("getAskPriceAndName", err)
		return
	}
	total := price * float64(quantity)
	user, err = getUser(c.id, c.db)
	if err != nil {
		log.Println("get user err", err)

		return
	}

	if float64(user.Balance) >= total {
		fmt.Println("Buy stock")
		user.Balance = user.Balance - int(total)
		err = addStock(&Stock{Quantity: int(quantity), Symbol: symbol, UserId: c.id, PricePaid: total, Company: company}, c.db)
		if err != nil {
			log.Println("get user err", err)
			return
		}
		err = updateUser(&user, c.db)
		if err != nil {
			log.Println("updateUser(&user):", err)
			return
		}
		c.sendMessage("Stock Added")

	} else {
		c.sendMessage("Not Enough Balance")

	}
}

func (c *Client) sell(symbol string, quantity float64) {
	//check if quantity available for sell
	var (
		stock   Stock
		err     error
		price   float64
		company string
		user    User
	)
	stock, err = getStock(&Stock{Symbol: symbol, UserId: c.id}, c.db)
	if err != nil {
		c.sendMessage("Stock Not available")
		return
	}
	if stock.Quantity < int(quantity) {
		c.sendMessage("Insufficient stocks available to sell")
		return

	}
	price, company, err = getSellPriceAndName(symbol)
	if err != nil {
		return
	}
	total := price * float64(quantity)
	user, err = getUser(c.id, c.db)
	if err != nil {
		c.sendMessage(fmt.Sprint("Error ", err))
		return
	}

	user.Balance = user.Balance + int(total)
	err = removeStock(&Stock{Quantity: int(quantity), Symbol: symbol, UserId: c.id, Company: company}, c.db)
	if err != nil {
		c.sendMessage(fmt.Sprint("Error ", err))
		return
	}
	err = updateUser(&user, c.db)
	if err != nil {
		c.sendMessage(fmt.Sprint("Error updateUser(&user) ", err))
		return
	}
	c.sendMessage("Stock sold")
}

func (c *Client) sendMessage(message string) {
	c.sendResponse(ResponseData{Typeofdata: "message", Data: message})

}

func (c *Client) sendResponse(response ResponseData) {
	c.conn.WriteJSON(response)

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
	c.db = getDb()
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

	client := &Client{conn: conn, send: make(chan []byte, 256), id: ""}
	client.register(token)
	go client.writePump()
	client.readPump()

}
