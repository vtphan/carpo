package main

import (
	"database/sql"
	"log"
	"time"
)

func (sub *Submission) SaveSubmission(studentID int) (id int, err error) {
	result, err := Database.Exec("insert into submission (problem_id, message, code, snapshot, student_id, status, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?)", sub.ProblemID, sub.Message, sub.Code, sub.Snapshot, studentID, sub.Status, sub.CreatedAt, sub.UpdatedAt)
	sid, _ := result.LastInsertId()
	return int(sid), nil
}

func (sub *Submission) UpdateSubmission(studentID int) (err error) {
	_, err = Database.Exec("update submission set message=?, code=?, status=?, updated_at=? where problem_id=? and student_id=?", sub.Message, sub.Code, sub.Status, sub.UpdatedAt, sub.ProblemID, sub.StudentID)
	return
}

func (sub *Submission) SetSubmissionStatus(status int) (err error) {

	_, err = Database.Exec("UPDATE submission SET status=?, updated_at=?  where id=?", status, time.Now(), sub.ID)

	if err != nil {
		log.Printf("SQL Error. Err: %v", err)
		log.Fatal(err)
	}

	log.Printf("Submission status set to %v for sub id: %v.\n", status, sub.ID)
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

func (sub *Submission) SetID(id int) {
	sub.ID = id

}
