package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var url = "http://data.benzinga.com/rest/richquoteDelayed?symbols="

func getQuote(sym string) (q interface{}, e error) {
	var (
		r    *http.Response
		body []byte
		i    map[string]interface{}
	)
	r, e = http.Get(url + sym)
	if e != nil {
		fmt.Println("http.Ge", e)
		return
	}

	defer r.Body.Close()
	body, e = ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("hutil.ReadA", e)

		return
	}

	e = json.Unmarshal(body, &i)

	q = i[sym]

	if e != nil {
		fmt.Println("on.Unmarsh", e)
		return
	}
	return
}

func getAskPriceAndName(symbol string) (f float64, name string, err error) {
	i, e := getQuote(symbol)
	if e != nil {
		fmt.Println("Error getAskPrice", e)
	}
	quote, ok := i.(map[string]interface{})
	if !ok {
		fmt.Println("break")
		return 0, "", fmt.Errorf("Parsing error")
	}
	return quote["askPrice"].(float64), quote["name"].(string), nil
}

func getSellPriceAndName(symbol string) (f float64, name string, err error) {
	i, e := getQuote(symbol)
	if e != nil {
		fmt.Println("Error getAskPrice", e)
	}
	quote, ok := i.(map[string]interface{})
	if !ok {
		fmt.Println("break")
		return 0, "", fmt.Errorf("Parsing error")
	}
	return quote["bidPrice"].(float64), quote["name"].(string), nil
}
