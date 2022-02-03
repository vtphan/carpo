package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
				fmt.Printf("Failed to Save Submission. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to save submission.",
					http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, string("Submission saved successfully."))

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

		// handle panic
		teacherID := int(body["teacher_id"].(float64))
		studentID := int(body["student_id"].(float64))
		score := int(body["score"].(float64))
		submissionID := int(body["submission_id"].(float64))

		switch r.Method {
		case http.MethodPost:
			_, err = AddScoreSQL.Exec(teacherID, studentID, submissionID, score, time.Now(), time.Now())
			if err != nil {

				fmt.Printf("Failed to Save Score. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to save Score.",
					http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, string("Submission graded successfully."))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}
