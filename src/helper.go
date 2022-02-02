package main

import (
	"encoding/json"
	"net/http"
)

func readRequestBody(r *http.Request) (req map[string]interface{}, err error) {

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return
}
