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

func teacherFeedbackHandler(w http.ResponseWriter, r *http.Request) {

	teacher_id := 0

	if user_id := r.Context().Value("user_id"); user_id != nil {
		teacher_id = user_id.(int)
	}

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

	sub := Submission{
		ID: f.SubmissionID,
	}
	graded, err := sub.IsGraded()

	f.TeacherID = teacher_id
	if err != nil {
		log.Printf("Failed to get grade status of the Submission . %+v Err. %v\n", sub, err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Failed to requeue submission.",
			http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Check the code block. if different that the submission, Update Feedback attributes in DB else add score only.
		var studentCode string
		rows, err := Database.Query("select code from submission where id = ?", f.SubmissionID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error querying db feedback save. Err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		for rows.Next() {
			rows.Scan(&studentCode)
		}

		hasFeedback := hasFeedbackOnCode(f.Code, studentCode)

		// if not graded and code doesnot have feedback, put back the submission in queue.
		if !graded && !hasFeedback {
			log.Printf("%v submission is not graded and has no feedback \n", f.SubmissionID)
			err = sub.SetSubmissionStatus(NewSub)
			if err != nil {
				log.Printf("Failed to requeue Submission . %v Err. %v\n", sub, err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to requeue submission.",
					http.StatusInternalServerError)
				return
			}

			// This msg statement is used to delete the notebook in instructor local machine.
			w.WriteHeader(http.StatusOK)
			resp := []byte(`{"msg": "Submission put back into the queue successfully."}`)
			fmt.Fprint(w, string(resp))
			return
		}

		// if graded and doesn't have feedback, give error.
		if graded && !hasFeedback {
			log.Printf("Failed to requeue Submission %v. Submission already graded and has no feedback. Err. %v\n", sub.ID, err)
			w.WriteHeader(http.StatusOK)
			resp := []byte(`{"msg": "This submission is already graded and has no feedback. It cannot go back into the submission queue."}`)
			fmt.Fprint(w, string(resp))
			return
		}

		// if has inline feedback, send feedback to student.
		if hasFeedback {
			log.Printf("%v feedback has inline feedback.\n", f.SubmissionID)
			_, err = Database.Exec("insert into grade (teacher_id, submission_id, student_id, score, code_feedback, comment, status, has_feedback, feedback_at, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", f.TeacherID, f.SubmissionID, f.StudnetID, 0, f.Code, f.Comment, 0, 1, time.Now(), time.Now(), time.Now())

			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Feedback already provided for %v. Updating...\n", f.SubmissionID)
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						_, err = Database.Exec("update grade set code_feedback=?, comment=?, has_feedback=1, feedback_at=?, status=0 where teacher_id=? and submission_id=?", f.Code, f.Comment, time.Now(), f.TeacherID, f.SubmissionID)
						// _, err = UpdateFeedbackSQL.Exec(f.Code, f.Comment, time.Now(), f.TeacherID, f.SubmissionID)
						if err != nil {
							log.Printf("Failed to update feedback %+v. Err: %v", f, err)
						}
						log.Printf("Feedback successfully updated.")
					}

				} else {
					log.Printf("Failed to save Feedback: %v Err. %v\n", f, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save Feedback.",
						http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Feedback is sent to the student."}`)
			fmt.Fprint(w, string(resp))
		}

	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	}

	// }

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

			log.Printf("Fetching Feedbacks for student_id %v... \n", student_id[0])

			f := Feedback{}
			rows, err := Database.Query("select grade.id, submission.problem_id, submission.message, code_feedback, comment, grade.updated_at from grade INNER JOIN submission on grade.submission_id = submission.id where grade.student_id = ? and grade.status = 0 and grade.has_feedback=1 order by grade.created_at desc", student_id[0])
			if err != nil {
				log.Printf("Error quering db for getSubmissionFeedbacks. Err: %v", err)
				return
			}

			defer rows.Close()

			for rows.Next() {
				rows.Scan(&f.ID, &f.ProblemID, &f.Message, &f.CodeFeedback, &f.Comment, &f.LastUpdatedAt)
				feedbacks = append(feedbacks, f)
			}

			// Set grade status to 1 which are sent to client
			// for _, feedback := range feedbacks {
			// 	stmt, err := Database.Prepare("update grade set status=?, updated_at=?  where id=?")
			// 	if err != nil {
			// 		log.Printf("SQL Error. Err: %v", err)
			// 	}
			// 	log.Printf("Set Feedback status to %v for Grade id: %v.\n", 1, feedback.ID)
			// 	_, err = stmt.Exec(1, time.Now(), feedback.ID)
			// }

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
