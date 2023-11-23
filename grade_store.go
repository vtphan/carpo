package main

import "time"

type GradeStore interface {
	SaveGradeFeedback(GradeFeedback) (int, error)
	UpdateGradeFeedback(GradeFeedback) error
	GetSubCodeFromID(int) (string, error)
	GetSubGrade(int) (GradeFeedback, error)
}

type GradeFeedback struct {
	ID           int       `json:"id"`
	TeacherID    int       `json:"teacher_id" db:"user_id"`
	SubmissionID int       `json:"submission_id"`
	Score        int       `json:"score"`
	Code         string    `json:"code"`
	Comment      string    `json:"comment"`
	Status       int       `json:"status"`
	HasFeedback  int       `json:"has_feedback"`
	FeedbackAt   time.Time `json:"feedback_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (db *Database) GetSubGrade(subID int) (g GradeFeedback, err error) {

	sqlStatement := `SELECT id, user_id, submission_id, score, code, comment, status, has_feedback, feedback_at FROM grades WHERE submission_id=$1 LIMIT 1;`

	row := db.DB.QueryRow(sqlStatement, subID)
	err = row.Scan(&g.ID, &g.TeacherID, &g.SubmissionID, &g.Score, &g.Code, &g.Comment, &g.Status, &g.HasFeedback, &g.FeedbackAt)

	if err != nil {
		return g, err
	}
	return
}

func (db *Database) SaveGradeFeedback(g GradeFeedback) (id int, err error) {

	sqlStatement := `INSERT into grades (user_id, submission_id, score, code, comment, status, has_feedback, feedback_at, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;`

	err = db.DB.QueryRow(sqlStatement, g.TeacherID, g.SubmissionID, g.Score, g.Code, g.Comment, g.Status, g.HasFeedback, g.FeedbackAt, g.CreatedAt, g.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	updateStatement := `UPDATE submissions set status=2 where id = $1`
	_, err = db.DB.Exec(updateStatement, g.SubmissionID)

	if err != nil {
		return id, err
	}

	return
}

func (db *Database) UpdateGradeFeedback(g GradeFeedback) (err error) {
	sqlStatement := `UPDATE grades set score=$1, code=$2, comment=$3, has_feedback=$4, feedback_at=$5, updated_at=$6 where user_id=$7 and submission_id=$8`

	_, err = db.DB.Exec(sqlStatement, g.Score, g.Code, g.Comment, g.HasFeedback, g.FeedbackAt, g.UpdatedAt, g.TeacherID, g.SubmissionID)

	if err != nil {
		return err
	}

	updateStatement := `UPDATE submissions set status=2 where id = $1`
	_, err = db.DB.Exec(updateStatement, g.SubmissionID)

	if err != nil {
		return err
	}

	return
}
