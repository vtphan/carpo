package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func problemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		student_id, ok := r.URL.Query()["student_id"]
		if !ok || len(student_id[0]) < 1 {
			log.Println("Url Param 'student_id' is missing")
			return
		}

		switch r.Method {
		case http.MethodGet:
			rows, err := Database.Query("select id, teacher_id, question from problem order by created_at desc limit 1")
			defer rows.Close()
			if err != nil {
				fmt.Errorf("Error quering db. Err: %v", err)
			}

			var (
				id, teacher_id int
				question       string
			)

			fmt.Printf("%v\n", rows)

			for rows.Next() {
				rows.Scan(&id, &teacher_id, &question)
			}
			resp := map[string]interface{}{
				"id":         id,
				"teacher_id": teacher_id,
				"question":   question,
			}

			_, err = AddStudentProblemStatusSQL.Exec(student_id[0], id, 0, time.Now(), time.Now())
			if err != nil {
				fmt.Printf("Failed to update student problem status(0) to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to update student problem status(0) to DB.",
					http.StatusInternalServerError)
				return
			}
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))

		case http.MethodPost:
			_, err := AddProblemSQL.Exec(body["teacher_id"], body["question"], time.Now(), time.Now())
			if err != nil {

				fmt.Printf("Failed to add question to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to add question to DB.",
					http.StatusInternalServerError)
				return

			}
			fmt.Printf("Added Problem: %v\n", body)
			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Question saved successfully."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}
