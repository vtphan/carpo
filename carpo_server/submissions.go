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
	rows, err := Database.Query("select created_at from submission where student_id=? and snapshot=2 order by created_at desc limit 1", sID)
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

		// fmt.Printf("Req body: %+v\n", body)

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
		sub_type, _ := strconv.Atoi(fmt.Sprintf("%v", body["snapshot"])) // 1: codesnapshot, 2: submission

		sub := Submission{
			ProblemID: pid,
			Message:   fmt.Sprintf("%v", body["message"]),
			Code:      fmt.Sprintf("%v", body["code"]),
			Snapshot:  sub_type,
			StudentID: studnet.ID,
			Status:    NewSub,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// fmt.Printf("Submission: %+v\n", sub)s

		key := fmt.Sprintf("%v-%v", studnet.ID, body["problem_id"])
		switch r.Method {
		case http.MethodPost:
			expiredProblem, _ := isExpired(pid)
			// Ignore snapshot if the problem is expired
			if expiredProblem && sub_type == 1 {
				log.Printf("Discard snapshot for inactive problem with key: %s. ", key)
				resp := []byte(`{"msg": "Snapshot no longer needed."}`)
				fmt.Fprint(w, string(resp))
				return
			}

			if !isAllowedSubmission(studnet.ID) && sub_type == 2 {
				log.Printf("Submission is not allowed within 30 seconds of previous submission. StudentID: %v\n", studnet.ID)
				resp := []byte(`{"msg": "Please wait for 30 seconds before you make another submission on this problem."}`)
				fmt.Fprint(w, string(resp))
				return
			}

			if val, ok := studentWorkSnapshot[key]; ok {
				// Check for codesnapshot
				if val.Code == sub.Code && sub_type == 1 {
					log.Printf("No change for key: %s.", key)
					resp := []byte(`{"msg": "No new change found."}`)
					fmt.Fprint(w, string(resp))
					return
				}
			}

			dbID, err := studnet.SaveSubmission(sub)
			if err != nil {
				log.Printf("Failed to Save Submission. %v Err. %v\n", sub, err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to save submission.",
					http.StatusInternalServerError)
				return
			}

			// Put SnapshotID for the studentWorkSnapshot that is saved to DB.
			sub.ID = dbID

			studentWorkSnapshot[key] = sub

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

func teacherSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	// return func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.Header().Set("Access-Control-Max-Age", "15")
	// role := "teacher"

	teacher_id := 0
	if user_id := r.Context().Value("user_id"); user_id != nil {
		teacher_id = user_id.(int)
	}

	if teacher_id == 0 {
		http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
		return
	}

	fmt.Printf("TeacherID: %v\n", teacher_id)

	submissions := make([]Submission, 0)

	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		// teacher_id, ok := query["id"]
		// if !ok || len(teacher_id) < 1 {
		// 	log.Printf("Url Param 'id' is missing.\n")
		// 	http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
		// 	return
		// }

		// teacher_name, ok := query["name"]
		// if !ok || len(teacher_name) < 1 {
		// 	log.Printf("Url Param 'name' is missing.\n")
		// 	http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
		// 	return
		// }
		sorting := "submission.created_at"
		if sort_by, ok := query["sort_by"]; ok {
			switch sort_by[0] {
			case "name":
				sorting = "student.name ASC"
			case "creation_time":
				sorting = "submission.created_at ASC"
			default:
				log.Printf("sort_by parameter is missing. Using default sort by created_at.\n")
			}
		}

		// Get name
		var name string
		rows, err := Database.Query("select name from teacher where id=?", teacher_id)
		defer rows.Close()
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(&name)
		}

		// if name != teacher_name[0] {
		// 	http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
		// 	return
		// }

		newSub := 0
		sqlSmt := `select count(*) from submission where status = 0`
		_ = Database.QueryRow(sqlSmt).Scan(&newSub)

		log.Printf("Fetching all submissions of students...\n")
		// Only Active problem
		sql := "select submission.id, message, code, student_id, name, problem_id, problem.format, submission.created_at, submission.updated_at from submission inner join student on submission.student_id = student.id and submission.status = 0 and submission.snapshot = 2 inner join problem on submission.problem_id = problem.id where problem.status = 1"
		// order by submission.created_at asc

		// combine the sorting option:
		sql = fmt.Sprintf("%s ORDER BY %s", sql, sorting)

		s := Submission{}
		rows, err = Database.Query(sql)
		defer rows.Close()
		if err != nil {
			log.Printf("Error querying db teacherSubmissionHandler. Err: %v", err)
			return
		}

		for rows.Next() {
			rows.Scan(&s.ID, &s.Message, &s.Code, &s.StudentID, &s.Name, &s.ProblemID, &s.Format, &s.CreatedAt, &s.UpdatedAt)

			// Add Previous grading of the student's submissions.
			// grades, _ := Database.Query("select submission.id, submission.created_at, problem.lifetime, grade.score, grade.created_at, grade.teacher_id from submission INNER join problem on submission.problem_id = problem.id left join grade on grade.submission_id = submission.id where submission.student_id = ? and problem.id = ? order by submission.created_at desc", s.StudentID, s.ProblemID)
			// s.Info = ""
			// for grades.Next() {
			// 	var scoreTime, subTime, problemLifeTime string
			// 	var score, teacher_id, sub_id int

			// 	grades.Scan(&sub_id, &subTime, &problemLifeTime, &score, &scoreTime, &teacher_id)
			// 	sTime, _ := time.Parse(time.RFC3339, subTime)
			// 	lifeTime, _ := time.Parse(time.RFC3339, problemLifeTime)
			// 	subMin := time.Now().Sub(sTime).Minutes()
			// 	timeLeft := lifeTime.Sub(sTime).Minutes()

			// 	if score != 0 {
			// 		gTime, _ := time.Parse(time.RFC3339, scoreTime)
			// 		s.Info += fmt.Sprintf("[ SubID %v | Submitted: %.1f minutes ago | Status: %v | Graded: %.1f minutes ago |  Time Left: %.1fm ] \n", sub_id, subMin, GradingMessage[score], time.Now().Sub(gTime).Minutes(), timeLeft)
			// 	} else {
			// 		s.Info += fmt.Sprintf("[ SubID %v | Submitted: %.1f minutes ago | Status: %v | Graded: %.1f minutes ago |  Time Left: %.1fm ] \n", sub_id, subMin, GradingMessage[score], 0.0, timeLeft)
			// 	}

			// 	s.Info += "\n\n"
			// }
			// s.Time = strconv.Itoa(s.CreatedAt.Hour()) + ":" + strconv.Itoa(s.CreatedAt.Minute())

			submissions = append(submissions, s)
		}

		// Set submission status to 1 which are sent to client
		// TODO: Uncomment this.
		// for _, subs := range submissions {
		// 	subs.SetSubmissionStatus(SubBeingLookedAt)
		// }

		if len(submissions) == 0 {
			log.Printf("No new submissions found.\n")
		}

		resp := Response{}
		resp.Remaining = newSub - len(submissions)
		sub, _ := json.Marshal(submissions)

		d := []map[string]interface{}{}
		_ = json.Unmarshal(sub, &d)
		resp.Data = d
		data, _ := json.Marshal(resp)
		fmt.Fprint(w, string(data))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// }
}

func submissionGradeHandler(w http.ResponseWriter, r *http.Request) {
	// return func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.Header().Set("Access-Control-Max-Age", "15")

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
	fmt.Printf("Req body : %s", jsonString)

	s := Grade{}
	json.Unmarshal(jsonString, &s)
	// fmt.Printf("S : %+v", s)

	s.TeacherID = teacher_id

	switch r.Method {
	case http.MethodPost:
		// Check the code block. if different that the submission, Update Feedback attributes in DB else add score only.
		var studentCode string
		rows, err := Database.Query("select code from submission where id = ?", s.SubmissionID)
		defer rows.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error querying db submissionGrade. Err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			rows.Scan(&studentCode)
		}

		if hasFeedbackOnCode(s.Code, studentCode) {
			_, err = AddFeedbackSQL.Exec(s.TeacherID, s.SubmissionID, s.StudnetID, s.Score, s.Code, s.Comment, 0, 1, time.Now(), time.Now(), time.Now())
		} else {
			_, err = AddScoreSQL.Exec(s.TeacherID, s.SubmissionID, s.StudnetID, s.Score, 0, time.Now(), time.Now())
		}

		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				log.Printf("Submission already graded for %v. Updating...\n", s.SubmissionID)
				if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
					if hasFeedbackOnCode(s.Code, studentCode) {
						_, err = UpdateScoreFeedbackSQL.Exec(s.Score, s.Code, s.Comment, time.Now(), s.TeacherID, s.SubmissionID)
					} else {
						_, err = UpdateScoreSQL.Exec(s.Score, time.Now(), s.SubmissionID)
					}

					if err != nil {
						log.Printf("Failed to update score %+v. Err: %v", s, err)
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
		err = sub.SetSubmissionStatus(SubGradedByTeacher)
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

func gradedSubmissionHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		gradedSubmissions := make([]GradedSubmission, 0)

		switch r.Method {
		case http.MethodGet:
			log.Printf("Fetching all graded submissions...\n")

			s := GradedSubmission{}
			rows, err := Database.Query("select submission.id, submission.message, submission.code, submission.created_at as sub_created_at, submission.student_id, grade.score, grade.created_at as grade_created_at, problem.id as problem_id, problem.lifetime, grade.comment from submission INNER join problem on submission.problem_id = problem.id left join grade on grade.submission_id = submission.id where grade.score in (1,2) order by submission.created_at desc")
			defer rows.Close()
			if err != nil {
				log.Printf("Error quering db gradedSubmissionHandler. Err: %v", err)
				return
			}

			for rows.Next() {
				rows.Scan(&s.ID, &s.Message, &s.Code, &s.SubCreatedAt, &s.StudentID, &s.Score, &s.GradedCreatedAt, &s.ProblemID, &s.ProblemLifeTime, &s.Comment)
				s.Time = strconv.Itoa(s.SubCreatedAt.Hour()) + ":" + strconv.Itoa(s.SubCreatedAt.Minute())
				gradedSubmissions = append(gradedSubmissions, s)
			}

			resp := Response{}

			sub, _ := json.Marshal(gradedSubmissions)

			d := []map[string]interface{}{}
			_ = json.Unmarshal(sub, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))
		}
	}

}
