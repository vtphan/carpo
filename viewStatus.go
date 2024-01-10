package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type StudentSubmissionStatus struct {
	ProblemID    int
	SubmissionID int
	Snapshot     int
	Code         string
	Submitted    string
	Score        int
	GradeAt      string
	HasFeedback  int
	Feedback     string
	FeedbackAt   string
}

type ProblemStatus struct {
	ProblemID     int
	Question      string
	Status        int
	PublishedDate time.Time
	LifeTime      time.Time
	ExpiresAt     string
	Solution      string
	UploadDate    time.Time
}

func viewStudentSubmissionStatus(db *sql.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		student_id, ok := c.GetQuery("student_id")
		if !ok || len(student_id) < 1 {
			log.Printf("Url Param 'student_id' is missing.\n")
			c.JSON(http.StatusUnauthorized, fmt.Sprintf("You are not authorized to view this status."))
			return
		}

		student_name, ok := c.GetQuery("student_name")
		if !ok || len(student_name) < 1 {
			log.Printf("Url Param 'student_name' is missing.\n")
			c.JSON(http.StatusUnauthorized, fmt.Sprintf("You are not authorized to view this status."))
			return
		}

		var name string
		rows, err := db.Query("select name from users where id=$1", student_id)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&name)
		}

		if name != student_name {
			c.JSON(http.StatusUnauthorized, fmt.Sprintf("You are not authorized to view this status."))
			return
		}

		// Get Submission status
		subStats := make([]StudentSubmissionStatus, 0)

		rows, err = db.Query("select s.problem_id, s.id, s.is_snapshot, s.code, s.created_at as created_at, grade.score, grade.updated_at, grade.has_feedback, grade.code, grade.feedback_at from submissions as s LEFT JOIN grades as grade on grade.submission_id = s.id where s.is_snapshot=2 and s.user_id = $1 Union select s.problem_id, s.id, s.is_snapshot, s.code, s.created_at, grade.score, grade.updated_at, grade.has_feedback, grade.code, grade.feedback_at from submissions as s INNER JOIN grades as grade on grade.submission_id = s.id where s.is_snapshot=1 and s.user_id = $2 order by created_at desc", student_id, student_id)
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

		rows, err = db.Query("select p.id, p.question, p.lifetime, p.status, p.created_at, s.code, s.broadcast, s.created_at from problems as p LEFT join solutions as s ON p.id = s.problem_id order by p.id desc")
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

		c.HTML(http.StatusOK, "stud_subs_status.tmpl", gin.H{"Name": name, "Stats": subStats, "PStats": problemStats})

	}
	return gin.HandlerFunc(fn)
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh:%dm", h, m)
}
