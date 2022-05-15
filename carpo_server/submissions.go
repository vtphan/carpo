package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
)

func isAllowedSubmission(sID int) bool {

	var prevSubAt string
	rows, err := Database.Query("select created_at from submission where student_id=? order by created_at desc limit 1", sID)
	defer rows.Close()
	if err != nil {
		log.Printf("Error SQL isAllowedSubmission. Error %v", err)
	}

	for rows.Next() {
		rows.Scan(&prevSubAt)
	}

	oldSubTime, _ := time.Parse(time.RFC3339, prevSubAt)

	return time.Now().Sub(oldSubTime).Seconds() >= 30.0
}

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

		id, err := studnet.GetIDFromName()
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "No Student found.",
				http.StatusNotFound)
			return
		}

		pid, _ := strconv.Atoi(fmt.Sprintf("%v", body["problem_id"]))

		sub := Submission{
			ProblemID: pid,
			Message:   fmt.Sprintf("%v", body["message"]),
			Code:      fmt.Sprintf("%v", body["code"]),
			StudentID: studnet.ID,
			Status:    NewSub,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		switch r.Method {
		case http.MethodPost:
			if !isAllowedSubmission(studnet.ID) {
				log.Printf("Submission is not allowed within 30 seconds of previous submission. StudentID: %v\n", studnet.ID)
				resp := []byte(`{"msg": "Please wait for 30 seconds before you make another submission on this problem."}`)
				fmt.Fprint(w, string(resp))
				return

			}
			_, err := studnet.SaveSubmission(sub)
			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Submission already found. Updating...")
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						studnet.UpdateSubmission(sub)
					}
				} else {

					log.Printf("Failed to Save Submission. %v Err. %v\n", sub, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save submission.",
						http.StatusInternalServerError)
					return
				}
			}
			_, err = AddStudentProblemStatusSQL.Exec(studnet.ID, pid, 1, time.Now(), time.Now())
			if err != nil {
				log.Printf("Failed to update student problem status (1) to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to update student problem status (1) to DB.",
					http.StatusInternalServerError)
				return
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
			newSub := 0
			sqlSmt := `select count(*) from submission where status = 0`
			_ = Database.QueryRow(sqlSmt).Scan(&newSub)

			log.Printf("Fetching submissions of students LIMIT 1...\n")

			s := Submission{}
			rows, err := Database.Query("select submission.id, message, code, student_id, name, problem_id, created_at, updated_at from submission inner join student on submission.student_id = student.id and submission.status = 0 order by created_at asc limit 1")
			defer rows.Close()
			if err != nil {
				log.Printf("Error quering db teacherSubmissionHandler. Err: %v", err)
				return
			}

			for rows.Next() {
				rows.Scan(&s.ID, &s.Message, &s.Code, &s.StudentID, &s.Name, &s.ProblemID, &s.CreatedAt, &s.UpdatedAt)

				// Add Previous grading of the student's submissions.
				grades, _ := Database.Query("select submission.id, submission.created_at, problem.lifetime, grade.score, grade.created_at, grade.teacher_id from submission INNER join problem on submission.problem_id = problem.id left join grade on grade.submission_id = submission.id where submission.student_id = ? and problem.id = ? order by submission.created_at desc", s.StudentID, s.ProblemID)
				s.Info = ""
				for grades.Next() {
					var scoreTime, subTime, problemLifeTime string
					var score, teacher_id, sub_id int

					grades.Scan(&sub_id, &subTime, &problemLifeTime, &score, &scoreTime, &teacher_id)
					sTime, _ := time.Parse(time.RFC3339, subTime)
					lifeTime, _ := time.Parse(time.RFC3339, problemLifeTime)
					subMin := time.Now().Sub(sTime).Minutes()
					timeLeft := lifeTime.Sub(sTime).Minutes()

					if score != 0 {
						gTime, _ := time.Parse(time.RFC3339, scoreTime)
						s.Info += fmt.Sprintf("[ SubID %v | Submitted: %.1f minutes ago | Status: %v | Graded: %.1f minutes ago |  Time Left: %.1f ] \n", sub_id, subMin, GradingMessage[score], time.Now().Sub(gTime).Minutes(), timeLeft)
					} else {
						s.Info += fmt.Sprintf("[ SubID %v | Submitted: %.1f minutes ago | Status: %v | Graded: %.1f minutes ago |  Time Left: %.1f ] \n", sub_id, subMin, GradingMessage[score], 0.0, timeLeft)
					}

					s.Info += "\n\n"
				}
				s.Time = strconv.Itoa(s.CreatedAt.Hour()) + ":" + strconv.Itoa(s.CreatedAt.Minute())

				submissions = append(submissions, s)
			}

			// Set submission status to 1 which are sent to client
			for _, subs := range submissions {
				subs.SetSubmissionStatus(SubBeingLookedAt)
			}

			if len(submissions) == 0 {
				log.Printf("No new submissions found.\n")
			}

			// fmt.Printf("%v", submissions)
			resp := Response{}
			if newSub >= 1 {
				resp.Remaining = newSub - 1
			}
			sub, _ := json.Marshal(submissions)

			d := []map[string]interface{}{}
			_ = json.Unmarshal(sub, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))
		case http.MethodPost:
			body, err := readRequestBody(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Error reading request body",
					http.StatusInternalServerError)
				return
			}

			subID, _ := strconv.Atoi(fmt.Sprintf("%v", body["submission_id"]))

			sub := Submission{
				ID: subID,
			}
			graded, err := sub.IsGraded()
			if graded {
				log.Printf("Failed to requeue Submission %v. Submission already graded. Err. %v\n", sub.ID, err)
				w.WriteHeader(http.StatusOK)
				resp := []byte(`{"msg": "This submission is already graded. It cannot go back into the submission queue."}`)
				fmt.Fprint(w, string(resp))
				return
			}

			err = sub.SetSubmissionStatus(NewSub)
			if err != nil {
				log.Printf("Failed to requeue Submission . %v Err. %v\n", sub, err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to requeue submission.",
					http.StatusInternalServerError)
				return

			}

			w.WriteHeader(http.StatusOK)
			resp := []byte(`{"msg": "Submission put back into the queue successfully."}`)
			fmt.Fprint(w, string(resp))

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

		s := Grade{}
		json.Unmarshal(jsonString, &s)

		switch r.Method {
		case http.MethodPost:
			_, err = AddScoreSQL.Exec(s.TeacherID, s.SubmissionID, s.StudnetID, s.Score, 1, time.Now(), time.Now())

			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Submission already graded for %v. Updating...\n", s.SubmissionID)
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						_, err := UpdateScoreSQL.Exec(s.Score, time.Now(), s.SubmissionID)
						if err != nil {
							log.Printf("Failed to update row %+v. Err: %v", s, err)
						}
						log.Printf("Score successfully updated.")
					}
				} else {
					log.Printf("Failed to Save Score %+v. Err. %v\n", s, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save Score.",
						http.StatusInternalServerError)

				}

			}

			sub := Submission{
				ID: s.SubmissionID,
			}
			err := sub.SetSubmissionStatus(SubGradedByTeacher)
			if err != nil {
				log.Printf("Failed to update Submission after grading submission. %v Err: %v\n", s, err)
			}

			_, err = AddStudentProblemStatusSQL.Exec(s.StudnetID, s.ProblemID, 2, time.Now(), time.Now())
			if err != nil {
				log.Printf("Failed to update student problem status (2) to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to update student problem status (2) to DB.",
					http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Submission graded successfully."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}
