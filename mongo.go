package main

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func connect() *mgo.Database {

	session, err := mgo.Dial(configuration.Db.IP + ":" + configuration.Db.Port)
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		panic(err)
	}
	return session.DB(configuration.Db.Name)

}

func getUser(id string) (u User, err error) {
	col := database.C(dbUserTable)
	err = col.Find(bson.M{"userid": id}).One(&u)
	if err != nil {
		log.Println("Err getUser:", err)
		return
	}
	return
}

func updateUser(user *User) error {
	change, err := database.C(dbUserTable).Upsert(bson.M{"userid": user.UserId}, user)
	if err != nil {
		return err
	}
	fmt.Println("User Balance updated", change)
	return nil
}

type Stock struct {
	Symbol    string
	Quantity  int
	UserId    string
	PricePaid float64
	Company   string
}

func getStock(st *Stock) (stock Stock, err error) {
	err = database.C(dbStockTable).Find(bson.M{"symbol": st.Symbol, "userid": st.UserId}).One(&stock)
	if err != nil {
		return
	}
	return
}

func addStock(newStock *Stock) error {
	oldStock, e := getStock(newStock)
	//stock found update
	if e == nil {
		newStock.Quantity = oldStock.Quantity + newStock.Quantity
	}
	change, err := database.C(dbStockTable).Upsert(bson.M{"symbol": newStock.Symbol, "userid": newStock.UserId}, newStock)
	if err != nil {
		return err
	}
	fmt.Println("Changes done", change)
	return nil
}

func removeStock(newStock *Stock) error {
	oldStock, e := getStock(newStock)
	//stock found update
	if e == nil {
		newStock.Quantity = oldStock.Quantity - newStock.Quantity
	}
	err := database.C(dbStockTable).Update(bson.M{"symbol": newStock.Symbol, "userid": newStock.UserId}, newStock)
	if err != nil {
		return err
	}
	return nil
}

func getPortfolio(id string) (stocks []Stock) {
	database.C(dbStockTable).Find(bson.M{"userid": id}).All(&stocks)
	return
}
