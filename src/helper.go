package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func readRequestBody(r *http.Request) (req map[string]interface{}, err error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &req)
		if err != nil {
			return nil, err
		}
	}
	return
}

func add(x, y int) int {
	return x + y
}
