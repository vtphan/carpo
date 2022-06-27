package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

// 15m 2d 45m
func getTimeLimit(timeLimit string) (val int, err error) {
	if strings.Contains(timeLimit, "m") || strings.Contains(timeLimit, "M") {
		tLimit := strings.Replace(timeLimit, "m", "", 1)
		tLimit = strings.Replace(tLimit, "M", "", 1)
		return strconv.Atoi(tLimit)

	} else if strings.Contains(timeLimit, "d") || strings.Contains(timeLimit, "D") {
		tLimit := strings.Replace(timeLimit, "d", "", 1)
		tLimit = strings.Replace(tLimit, "D", "", 1)
		tDay, err := strconv.Atoi(tLimit)
		if err != nil {
			return 0, err
		}
		return tDay * 1440, nil

	} else {
		return 0, fmt.Errorf("Error: Invalid value %v for time_limit.", timeLimit)
	}
}

func hasFeedbackOnCode(codeFromT, codeFromS string) bool {
	// Skip first line from TeacherCode.
	teacherCode := strings.Join(strings.Split(codeFromT, "\n")[1:], "\n")
	if strings.EqualFold(strings.Replace(teacherCode, " ", "", -1), strings.Replace(codeFromS, " ", "", -1)) {
		return false
	}
	return true
}
