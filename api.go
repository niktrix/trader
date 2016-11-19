package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var url = "http://data.benzinga.com/rest/richquoteDelayed?symbols="

func getQuote(sym string) (q interface{}, err error) {
	var (
		r    *http.Response
		body []byte
		i    map[string]interface{}
	)
	r, err = http.Get(url + sym)
	if err != nil {
		return
	}
	defer r.Body.Close()
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &i)
	q = i[sym]
	if err != nil {
		return
	}
	return
}

func getAskPriceAndName(symbol string) (f float64, name string, err error) {
	var i interface{}
	i, err = getQuote(symbol)
	if err != nil {
		return
	}
	quote, ok := i.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Parsing error")
		return
	}
	return quote["askPrice"].(float64), quote["name"].(string), nil
}

func getSellPriceAndName(symbol string) (f float64, name string, err error) {
	var i interface{}
	i, err = getQuote(symbol)
	if err != nil {
		return
	}
	quote, ok := i.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Parsing error")
		return
	}
	return quote["bidPrice"].(float64), quote["name"].(string), nil
}
