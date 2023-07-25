package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
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
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&name)
		}

		if name != student_name[0] {
			http.Error(w, fmt.Sprintf("You are not authorized to view this status."), http.StatusUnauthorized)
			return
		}

		// Get Submission status
		subStats := make([]StudentSubmissionStatus, 0)

		rows, err = Database.Query("select submission.problem_id, submission.id, submission.snapshot, submission.code, submission.created_at, grade.score, grade.updated_at, grade.has_feedback, grade.code_feedback, grade.feedback_at from submission LEFT JOIN grade on grade.submission_id = submission.id where submission.snapshot=2 and submission.student_id = ? Union select submission.problem_id, submission.id, submission.snapshot, submission.code, submission.created_at, grade.score, grade.updated_at, grade.has_feedback, grade.code_feedback, grade.feedback_at from submission INNER JOIN grade on grade.submission_id = submission.id where submission.snapshot=1 and submission.student_id = ? order by submission.created_at desc", student_id[0], student_id[0])
		if err != nil {
			log.Printf("Error quering db viewStudentSubmissionStatus. Err: %v", err)
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			stat := StudentSubmissionStatus{}
			var (
				SubCreatedAt      string
				GradeCreatedAt    string
				FeedbackCreatedAt string
				score             sql.NullInt64
			)
			rows.Scan(&stat.ProblemID, &stat.SubmissionID, &stat.Snapshot, &stat.Code, &SubCreatedAt, &score, &GradeCreatedAt, &stat.HasFeedback, &stat.Feedback, &FeedbackCreatedAt)
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

		// Fetch Problem and solutions
		problemStats := make([]ProblemStatus, 0)

		rows, err = Database.Query("select problem.id, problem.question, problem.lifetime, problem.status, problem.created_at, solution.code, solution.broadcast, solution.created_at from problem LEFT join solution ON problem.id = solution.problem_id order by problem.id desc")
		if err != nil {
			log.Printf("Error quering db ProblemStatus. Err: %v", err)
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				PCreatedAt    string
				PDeadlineAt   string
				isBroadcasted int64
			)
			stat := ProblemStatus{}

			rows.Scan(&stat.ProblemID, &stat.Question, &PDeadlineAt, &stat.Status, &PCreatedAt, &stat.Solution, &isBroadcasted, &stat.UploadDate)
			if isBroadcasted == 0 {
				stat.Solution = ""
			}
			stime, _ := time.Parse(time.RFC3339, PDeadlineAt)
			stat.LifeTime = stime
			// stat.LifeTime = fmt.Sprintf("%s ago", fmtDuration(time.Now().Sub(stime)))

			stime, _ = time.Parse(time.RFC3339, PCreatedAt)
			stat.PublishedDate = stime

			problemStats = append(problemStats, stat)
		}

		data := struct {
			Name   string
			Stats  []StudentSubmissionStatus
			PStats []ProblemStatus
		}{
			Name:   name,
			Stats:  subStats,
			PStats: problemStats,
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

func problemStatus(rows *sql.Rows, problemOnWatch map[int]int) (pGradeStat ProblemGradeStatus) {

	var (
		ungraded, correct, incorrect sql.NullInt64
	)

	rows.Scan(&pGradeStat.ProblemID, &pGradeStat.Question, &pGradeStat.PublishedDate, &pGradeStat.LifeTime, &pGradeStat.ProblemStatus, &ungraded, &correct, &incorrect, &pGradeStat.SolutionID, &pGradeStat.Solution)

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

	pGradeStat.OnWatch = problemOnWatch[int(pGradeStat.ProblemID)]

	return

}

func viewProblemStatus(w http.ResponseWriter, r *http.Request) {
	ids := []int{}
	problemOnWatch := map[int]int{}

	// Get On Watch for the Problem
	rows, err := Database.Query(fmt.Sprintf("select problem_id, count(*) as on_watch from watched group by problem_id"))

	if err != nil {
		log.Printf("Error quering db getProblems for On Watch. Err: %v", err)
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			problem_id, on_watch int64
		)
		rows.Scan(&problem_id, &on_watch)
		problemOnWatch[int(problem_id)] = int(on_watch)

	}

	// Get Problem Grading Status
	pGradeStats := make([]ProblemGradeStatus, 0)
	rows, err = Database.Query("select submission.problem_id, problem.question, problem.created_at, problem.lifetime, problem.status, sum(case when submission.status in (0,1) and submission.snapshot=2 then 1 end) as Ungraded, sum(case when grade.score = 1 then 1 end) as Correct, sum(case when grade.score = 2 then 1 end) as Incorrect, s.id, s.code from problem LEFT join submission on problem.id = submission.problem_id LEFT join grade on submission.id = grade.submission_id LEFT join solution as s on problem.id = s.problem_id group by submission.problem_id order by submission.problem_id desc")
	if err != nil {
		log.Printf("Error quering db viewProblemStatus. Err: %v", err)
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		pGradeStat := problemStatus(rows, problemOnWatch)

		pGradeStat.ExpiresAt = fmt.Sprintf("To be due in %s", fmtDuration(pGradeStat.LifeTime.Sub(time.Now())))
		pGradeStats = append(pGradeStats, pGradeStat)
		ids = append(ids, pGradeStat.ProblemID)
	}

	// Array of int to string with ,
	IDs := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	// Get Problems that don't have submissions yet.
	rows, err = Database.Query(fmt.Sprintf("select p.id, p.question, p.created_at, p.lifetime, p.status, s.id, s.code from problem as p left join solution as s on p.id = s.problem_id where p.id not in (%s) order by p.id desc", IDs))

	if err != nil {
		log.Printf("Error quering db getProblems. Err: %v", err)
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		pGradeStat := ProblemGradeStatus{}
		rows.Scan(&pGradeStat.ProblemID, &pGradeStat.Question, &pGradeStat.PublishedDate, &pGradeStat.LifeTime, &pGradeStat.ProblemStatus, &pGradeStat.SolutionID, &pGradeStat.Solution)
		pGradeStat.ExpiresAt = fmt.Sprintf("To be due in %s", fmtDuration(pGradeStat.LifeTime.Sub(time.Now())))
		pGradeStats = append(pGradeStats, pGradeStat)
	}

	// Sort the merged array of struct by ProblemID
	sort.Slice(pGradeStats, func(i, j int) bool {
		return pGradeStats[i].ProblemID > pGradeStats[j].ProblemID
	})

	resp := Response{}
	sub, _ := json.Marshal(pGradeStats)

	d := []map[string]interface{}{}
	_ = json.Unmarshal(sub, &d)
	resp.Data = d
	da, _ := json.Marshal(resp)
	fmt.Fprint(w, string(da))
	return

}

func problemOnWatch(w http.ResponseWriter, r *http.Request) {

	problem_id, ok := r.URL.Query()["problem_id"]
	if !ok || len(problem_id[0]) < 1 {
		log.Println("Url Param 'problem_id' is missing")
		http.Error(w, fmt.Sprintf("problem_id is missing."), http.StatusUnprocessableEntity)
		return
	}
	pWatch := []ProblemWatch{}
	// Get On Watch for the Problem
	rows, err := Database.Query("select w.id as id, w.problem_id, w.submission_id, s.code, w.reason from watched as w inner join submission as s on w.submission_id = s.id where w.problem_id = ?;", problem_id[0])

	if err != nil {
		log.Printf("Error quering db problemOnWatch. Err: %v", err)
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		watch := ProblemWatch{}
		rows.Scan(&watch.ID, &watch.ProblemID, &watch.SubmissionID, &watch.Code, &watch.Reason)
		pWatch = append(pWatch, watch)
	}

	// fmt.Printf("ProblemWatch: %+v\n", pWatch)

	resp := Response{}
	sub, _ := json.Marshal(pWatch)

	d := []map[string]interface{}{}
	_ = json.Unmarshal(sub, &d)
	resp.Data = d
	da, _ := json.Marshal(resp)
	fmt.Fprint(w, string(da))
	return

}
