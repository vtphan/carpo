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

func studentSubmissionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		name := fmt.Sprintf("%v", body["name"])
		studnet := Studnet{
			Name: name,
		}
		fmt.Printf("Studnet: %v\n", studnet)
		id, err := studnet.GetIDFromName()
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "No Student found.",
				http.StatusNotFound)
			return
		}

		sub := Submission{
			// TODO: question id should be in request body.
			QuestionID: 100,
			Message:    fmt.Sprintf("%v", body["message"]),
			// Message:    body["message"].(string),
			Code:      fmt.Sprintf("%v", body["code"]),
			StudentID: studnet.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		switch r.Method {
		case http.MethodPost:
			// TODO: Do we allow duplicate submissions from same student ?
			_, err := studnet.SaveSubmission(sub)
			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Submission already found. Updating...")
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						studnet.UpdateSubmission(sub)
					}
				} else {

					fmt.Printf("Failed to Save Submission. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save submission.",
						http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Submission saved successfully."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}

func teacherSubmissionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// role := "teacher"
		submissions := make([]Submission, 0)

		switch r.Method {
		case http.MethodGet:
			s := Submission{}
			rows, _ := Database.Query("select id, message, code, student_id, question_id, created_at,updated_at from submission order by updated_at desc")
			defer rows.Close()

			for rows.Next() {
				rows.Scan(&s.ID, &s.Message, &s.Code, &s.StudentID, &s.QuestionID, &s.CreatedAt, &s.UpdatedAt)
				submissions = append(submissions, s)
			}

			// fmt.Printf("%v", submissions)
			resp := Response{}
			sub, _ := json.Marshal(submissions)

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

func submissionGradeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		jsonString, _ := json.Marshal(body)

		s := Score{}
		json.Unmarshal(jsonString, &s)

		switch r.Method {
		case http.MethodPost:
			_, err = AddScoreSQL.Exec(s.TeacherID, s.StudnetID, s.SubmissionID, s.Score, time.Now(), time.Now())

			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Submission already graded. Updating...")
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						_, err := UpdateScoreSQL.Exec(s.Score, time.Now(), s.TeacherID, s.SubmissionID)
						if err != nil {
							log.Printf("Failed to update row %+v. Err: %v", s, err)
						}
						log.Printf("Score successfully updated.")
					}
				} else {
					fmt.Printf("Failed to Save Score. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save Score.",
						http.StatusInternalServerError)

				}

			}

			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Submission graded successfully."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}
