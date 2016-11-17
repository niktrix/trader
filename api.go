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
