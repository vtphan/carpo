package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func viewStudentSubmissionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		student_id, ok := query["student_id"]
		if !ok || len(student_id) < 1 {
			log.Printf("Url Param 'student_id' is missing.\n")
			http.Error(w, fmt.Sprintf("You are not authorized to view this status."), http.StatusUnauthorized)
			return
		}

		student_name, ok := query["student_name"]
		if !ok || len(student_name) < 1 {
			log.Printf("Url Param 'student_name' is missing.\n")
			http.Error(w, fmt.Sprintf("You are not authorized to view this status."), http.StatusUnauthorized)
			return
		}

		// Get name
		var name string
		rows, err := Database.Query("select name from student where id=?", student_id[0])
		defer rows.Close()
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(&name)
		}

		if name != student_name[0] {
			http.Error(w, fmt.Sprintf("You are not authorized to view this status."), http.StatusUnauthorized)
			return
		}

		// Get Submission status
		subStats := make([]StudentSubmissionStatus, 0)

		rows, err = Database.Query("select submission.problem_id, submission.id, submission.created_at, grade.score, grade.updated_at, grade.has_feedback, grade.feedback_at from submission LEFT JOIN grade on submission.id = grade.submission_id where submission.student_id = ? order by submission.created_at desc", student_id[0])

		defer rows.Close()
		if err != nil {
			log.Printf("Error quering db viewStudentSubmissionStatus. Err: %v", err)
		}

		for rows.Next() {
			stat := StudentSubmissionStatus{}
			var (
				SubCreatedAt      string
				GradeCreatedAt    string
				FeedbackCreatedAt string
				score             sql.NullInt64
			)
			rows.Scan(&stat.ProblemID, &stat.SubmissionID, &SubCreatedAt, &score, &GradeCreatedAt, &stat.HasFeedback, &FeedbackCreatedAt)
			if !score.Valid {
				score.Int64 = 0
			}
			stat.Score = int(score.Int64)

			stime, _ := time.Parse(time.RFC3339, SubCreatedAt)
			stat.Submitted = fmt.Sprintf("%s ago", fmtDuration(time.Now().Sub(stime)))

			stat.GradeAt = ""
			ftime, _ := time.Parse(time.RFC3339, GradeCreatedAt)
			if stat.Score == 1 || stat.Score == 2 {
				stat.GradeAt = fmt.Sprintf("%s ago", fmtDuration(time.Now().Sub(ftime)))
			}

			stat.FeedbackAt = ""
			gtime, _ := time.Parse(time.RFC3339, FeedbackCreatedAt)
			if stat.HasFeedback == 1 {
				stat.FeedbackAt = fmt.Sprintf("%s ago", fmtDuration(time.Now().Sub(gtime)))
			}

			subStats = append(subStats, stat)
		}

		data := struct {
			Name  string
			Stats []StudentSubmissionStatus
		}{
			Name:  name,
			Stats: subStats,
		}

		t, err := template.New("").Funcs(template.FuncMap{"add": add}).Parse(STUDENT_SUBMISSION_STATUS_TEMPLATE)
		if err != nil {
			log.Printf("%v\n", err)
		}

		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("%v\n", err)
		}

	}
}

func problemStatus(rows *sql.Rows) (pGradeStat ProblemGradeStatus) {

	var (
		ungraded, correct, incorrect sql.NullInt64
	)

	rows.Scan(&pGradeStat.ProblemID, &pGradeStat.PublishedDate, &pGradeStat.ProblemStatus, &pGradeStat.UnpublishedDated, &ungraded, &correct, &incorrect)

	if !ungraded.Valid {
		ungraded.Int64 = 0
	}
	pGradeStat.Ungraded = int(ungraded.Int64)

	if !correct.Valid {
		correct.Int64 = 0
	}
	pGradeStat.Correct = int(correct.Int64)

	if !incorrect.Valid {
		incorrect.Int64 = 0
	}
	pGradeStat.Incorrect = int(incorrect.Int64)

	return

}

func viewProblemStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Problem Grading Status
		pGradeStats := make([]ProblemGradeStatus, 0)
		rows, err := Database.Query("select submission.problem_id, problem.created_at, problem.status, problem.updated_at, sum(case when submission.status in (0,1) then 1 end) as Ungraded, sum(case when grade.score = 1 then 1 end) as Correct, sum(case when grade.score = 2 then 1 end) as Incorrect from submission LEFT join grade on submission.id = grade.submission_id  INNER join problem on problem.id = submission.problem_id group by problem_id order by problem_id desc")

		defer rows.Close()
		if err != nil {
			log.Printf("Error quering db viewProblemStatus. Err: %v", err)
		}

		for rows.Next() {
			pGradeStat := problemStatus(rows)
			pGradeStats = append(pGradeStats, pGradeStat)
		}

		data := struct {
			Stats []ProblemGradeStatus
		}{
			Stats: pGradeStats,
		}

		t, err := template.New("").Parse(PROBLEM_GRADE_STATUS_TEMPLATE)
		if err != nil {
			log.Printf("%v\n", err)
		}

		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("%v\n", err)
		}

	}

}

func problemDetail() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		problem_id, ok := query["problem_id"]
		if !ok || len(problem_id) < 1 {
			log.Printf("Url Param 'problem_id' is missing.\n")
			http.Error(w, fmt.Sprintf("Invalid Problem Id."), http.StatusUnauthorized)
			return
		}

		var (
			problem    string
			pGradeStat ProblemGradeStatus
		)

		rows, err := Database.Query("select question from problem where id = ?", problem_id[0])
		defer rows.Close()
		if err != nil {
			log.Printf("Error quering db problemQuestion. Err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			rows.Scan(&problem)
		}

		rows, err = Database.Query("select submission.problem_id, problem.created_at, problem.status, problem.updated_at, sum(case when submission.status in (0,1) then 1 end) as Ungraded, sum(case when grade.score = 1 then 1 end) as Correct, sum(case when grade.score = 2 then 1 end) as Incorrect from submission LEFT join grade on submission.id = grade.submission_id  INNER join problem on problem.id = submission.problem_id where problem.id =? group by problem_id order by problem_id desc", problem_id[0])

		defer rows.Close()
		if err != nil {
			log.Printf("Error quering db problemDetail. Err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			pGradeStat = problemStatus(rows)
		}

		data := struct {
			Stats    ProblemGradeStatus
			Question string
		}{
			Stats:    pGradeStat,
			Question: strings.TrimSpace(problem),
		}

		t, err := template.New("").Parse(PROBLEM_DETAIL_TEMPLATE)
		if err != nil {
			log.Printf("%v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, data)
		if err != nil {
			log.Printf("%v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}

}
