package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
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

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh:%dm", h, m)
}
