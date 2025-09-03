package main

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

// NullableString represents a nullable string value.
type NullableString struct {
	sql.NullString
}

type FeedbackStore interface {
	GetStudentFeedbackFromPID(int, int) ([]StudentPidFeedback, error)
}

type StudentPidFeedback struct {
	ID          int       `json:"id"`
	StudentID   int       `json:"user_id" db:"user_id"`
	ProblemID   int       `json:"problem_id"`
	Code        string    `json:"code"`
	Score       int       `json:"score"`
	GCode       string    `json:"gcode"`
	HasFeedback int       `json:"has_feedback"`
	FeedbackAt  time.Time `json:"feedback_at"`
	Rating      int       `json:"rating"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

func (db *Database) GetStudentFeedbackFromPID(sID int, pID int) ([]StudentPidFeedback, error) {

	// Get Submission status
	feedbackStats := make([]StudentPidFeedback, 0)

	sqlStatement := `select g.id, g.code, COALESCE(g.has_feedback,0), COALESCE(g.rating,0), COALESCE(g.feedback_at, '2025-07-28 13:59:54.538388-05'), g.score, s.code, g.created_at from grades as g INNER JOIN submissions as s ON g.submission_id = s.id where s.user_id = $1 and s.problem_id = $2 order by g.created_at desc; 
`
	rows, err := db.DB.Query(sqlStatement, sID, pID)
	if err != nil {
		log.Printf("Error quering db viewStudentSubmissionStatus. Err: %v", err)
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		stat := StudentPidFeedback{}
		rows.Scan(&stat.ID, &stat.Code, &stat.HasFeedback, &stat.Rating, &stat.FeedbackAt, &stat.Score, &stat.GCode, &stat.CreatedAt)
		// &_hasfeedbac, &_gcode, &_gfeedback)
		stat.StudentID = sID
		stat.ProblemID = pID

		feedbackStats = append(feedbackStats, stat)

	}

	return feedbackStats, nil
}
