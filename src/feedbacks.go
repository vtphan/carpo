package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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
			_, err = UpdateFeedbackSQL.Exec(f.Code, f.Comment, time.Now(), f.TeacherID, f.SubmissionID)

			if err != nil {

				fmt.Printf("Failed to save Feedback: %v Err. %v\n", f, err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to save Feedback.",
					http.StatusInternalServerError)

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

		question_id, ok := r.URL.Query()["question_id"]
		if !ok || len(question_id[0]) < 1 {
			log.Println("Url Param 'question_id' is missing")
			return
		}
		fmt.Printf("%v", question_id)

		feedbacks := make([]Feedback, 0)

		switch r.Method {
		case http.MethodGet:
			f := Feedback{}
			rows, err := Database.Query("select student.name, submission.message, code_feedback, comment, grade.updated_at  from grade INNER JOIN submission on grade.submission_id = submission.id INNER JOIN student on submission.student_id = student.id where submission.problem_id = ?", question_id[0])
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
