package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func viewStudentSubmissionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		student_id, ok := r.URL.Query()["student_id"]
		if !ok || len(student_id[0]) < 1 {
			log.Printf("Url Param 'student_id' is missing.\n")
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

		// Get Submission status
		subStats := make([]StudentSubmissionStatus, 0)
		stat := StudentSubmissionStatus{}
		rows, err = Database.Query("select submission.problem_id, submission.created_at, grade.score, grade.created_at from grade INNER JOIN submission on grade.submission_id = submission.id where grade.student_id = ? order by submission.created_at desc", student_id[0])

		defer rows.Close()
		if err != nil {
			fmt.Printf("Error quering db. Err: %v", err)
		}

		for rows.Next() {
			var (
				SubCreatedAt, GradeCreatedAt string
			)
			rows.Scan(&stat.ProblemID, &SubCreatedAt, &stat.Score, &GradeCreatedAt)
			stime, _ := time.Parse(time.RFC3339, SubCreatedAt)
			gtime, _ := time.Parse(time.RFC3339, GradeCreatedAt)
			stat.Submitted = fmt.Sprintf("%.1f min ago", time.Now().Sub(stime).Minutes())
			stat.Graded = fmt.Sprintf("%.1f min ago", time.Now().Sub(gtime).Minutes())
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

func viewProblemStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Problem Grading Status
		pGradeStats := make([]ProblemGradeStatus, 0)
		pGradeStat := ProblemGradeStatus{}
		rows, err := Database.Query("select problem_id, grade.score, count(*) from submission LEFT join grade on submission.id = grade.submission_id group by submission.problem_id,grade.score order by submission.problem_id desc")

		defer rows.Close()
		if err != nil {
			fmt.Printf("Error quering db. Err: %v", err)
		}

		for rows.Next() {

			rows.Scan(&pGradeStat.ProblemID, &pGradeStat.Score, &pGradeStat.Count)
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
