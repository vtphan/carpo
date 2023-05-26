package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func showSnapshotHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(studentWorkSnapshot)
		fmt.Fprint(w, string(data))
	}

}

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

// CORS Middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set headers
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Println("ok")

		// Next
		next.ServeHTTP(w, r)
		return
	})
}

func AuthorizeTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		fmt.Printf("Auth HEADER: %v\n", authHeader)
		w.Header().Add("Connection", "keep-alive")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS, GET, DELETE, PUT")
		w.Header().Add("Access-Control-Allow-Headers", "Authorization, content-type")
		w.Header().Add("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if len(authHeader) != 2 {
			fmt.Println("Malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Malformed Token"))
			return
		} else {
			token := authHeader[1]
			id := 0
			name := ""
			rows, err := Database.Query("select id, name from teacher where uuid = ?", token)
			defer rows.Close()
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			for rows.Next() {
				rows.Scan(&id, &name)
			}
			if id == 0 && name == "" {
				fmt.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}
			w.WriteHeader(http.StatusOK)
			ctx := context.WithValue(r.Context(), "user_id", id)
			ctx = context.WithValue(ctx, "user_name", name)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
