package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mattn/go-sqlite3"
)

func teacherFeedbackHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		jsonString, _ := json.Marshal(body)

		f := Grade{}
		json.Unmarshal(jsonString, &f)

		switch r.Method {
		case http.MethodPost:
			_, err = AddFeedbackSQL.Exec(f.TeacherID, f.SubmissionID, f.StudnetID, 0, f.Code, f.Comment, time.Now(), time.Now())

			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Feedback already provided for %v. Updating...\n", f.SubmissionID)
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						_, err = UpdateFeedbackSQL.Exec(f.Code, f.Comment, time.Now(), f.TeacherID, f.SubmissionID)
						if err != nil {
							log.Printf("Failed to update feedback %+v. Err: %v", f, err)
						}
						log.Printf("Feedback successfully updated.")
					}

				} else {
					fmt.Printf("Failed to save Feedback: %v Err. %v\n", f, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save Feedback.",
						http.StatusInternalServerError)

				}

			}

			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Feedback saved successfully."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		}

	}

}

func getSubmissionFeedbacks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		student_id, ok := r.URL.Query()["student_id"]
		if !ok || len(student_id[0]) < 1 {
			log.Println("Url Param 'student_id' is missing")
			return
		}

		feedbacks := make([]Feedback, 0)

		switch r.Method {
		case http.MethodGet:
			// Get Active question_id
			question_id := 0
			rows, err := Database.Query("select id from problem order by id desc limit 1")
			if err != nil {
				fmt.Printf("Error quering db. Err: %v", err)
			}
			for rows.Next() {
				rows.Scan(&question_id)
			}

			fmt.Printf("Fetching Feedbacks for student_id %v on question_id %v...\n", student_id[0], question_id)

			f := Feedback{}
			rows, err = Database.Query("select student.name, student.id, code_feedback, comment, grade.updated_at from grade INNER JOIN submission on grade.submission_id = submission.id INNER JOIN student on grade.student_id = student.id where student.id = ? and submission.problem_id = ? order by grade.created_at desc", student_id[0], question_id)

			// where submission.problem_id = ?", student_id[0]
			defer rows.Close()
			if err != nil {
				fmt.Printf("Error quering db. Err: %v", err)
			}

			for rows.Next() {
				rows.Scan(&f.Name, &f.Message, &f.CodeFeedback, &f.Comment, &f.LastUpdatedAt)
				feedbacks = append(feedbacks, f)
			}

			resp := Response{}
			sub, _ := json.Marshal(feedbacks)
			d := []map[string]interface{}{}
			_ = json.Unmarshal(sub, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}
