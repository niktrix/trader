package main

import (
	"gopkg.in/mgo.v2"
)

func connect() *mgo.Database {

	session, err := mgo.Dial(configuration.Db.IP + ":" + configuration.Db.Port)
	if err != nil {
		panic(err)
	}
	return session.DB(configuration.Db.Name)

}
