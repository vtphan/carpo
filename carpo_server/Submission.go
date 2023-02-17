package main

import (
	"database/sql"
	"log"
	"time"
)

func (sub *Submission) SaveSubmission(studentID int) (id int, err error) {
	var result sql.Result
	result, err = AddSubmissionSQL.Exec(sub.ProblemID, sub.Message, sub.Code, sub.Snapshot, studentID, sub.Status, sub.CreatedAt, sub.UpdatedAt)
	sid, _ := result.LastInsertId()
	return int(sid), nil
}

func (sub *Submission) UpdateSubmission(studentID int) (err error) {
	_, err = UpdateSubmissionSQL.Exec(sub.Message, sub.Code, sub.Status, sub.UpdatedAt, sub.ProblemID, sub.StudentID)
	return
}

func (sub *Submission) SetSubmissionStatus(status int) (err error) {
	stmt, err := Database.Prepare("update submission set status=?, updated_at=?  where id=?")
	if err != nil {
		log.Printf("SQL Error %v. Err: %v", stmt, err)
	}
	log.Printf("Submission status set to %v for sub id: %v.\n", status, sub.ID)
	_, err = stmt.Exec(status, time.Now(), sub.ID)
	return
}

func (sub *Submission) IsGraded() (graded bool, err error) {
	score := 0
	sqlSmt := `select score from grade where submission_id=?`
	err = Database.QueryRow(sqlSmt, sub.ID).Scan(&score)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}
	if score == 1 || score == 2 {
		return true, nil
	}
	return

}
