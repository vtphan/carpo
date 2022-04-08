package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func (sub *Submission) SaveSubmission(studentID int) (id int, err error) {
	var result sql.Result
	result, err = AddSubmissionSQL.Exec(sub.ProblemID, sub.Message, sub.Code, studentID, sub.Status, sub.CreatedAt, sub.UpdatedAt)
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
		log.Printf("SQL Error. Err: %v", err)
	}
	fmt.Printf("Submission status set to %v for sub id: %v.\n", status, sub.ID)
	_, err = stmt.Exec(status, time.Now(), sub.ID)
	return
}
