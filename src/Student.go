package main

import (
	"database/sql"
	"fmt"
	"log"
)

type Studnet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (st *Studnet) Add() (msg string, alreadyExists bool, err error) {

	rows, err := Database.Query("select name from student where name=?", st.Name)
	defer rows.Close()
	if err != nil {
		return "", false, err
	}

	for rows.Next() {
		return fmt.Sprintf("User %v already exists as %v.", st.Name, "student"), true, nil
	}
	_, err = AddStudentSQL.Exec(st.Name)
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("User %v created as %v.", st.Name, "student"), false, nil

}

func (st *Studnet) SaveSubmission(s Submission) (id int, err error) {
	var result sql.Result
	result, err = AddSubmissionSQL.Exec(s.QuestionID, s.Message, s.Code, st.ID, s.CreatedAt, s.UpdatedAt)
	if err != nil {

		return 0, err
	}

	sid, _ := result.LastInsertId()
	return int(sid), nil

}

func (st *Studnet) UpdateSubmission(s Submission) {

	_, err := UpdateSubmissionSQL.Exec(s.Message, s.Code, s.UpdatedAt, s.QuestionID, s.StudentID)
	if err != nil {
		log.Printf("Failed to update row %+v. Err: %v", s, err)
	}
	log.Printf("Submission successfully updated.")
}

func (st *Studnet) GetIDFromName() (id int, err error) {

	rows, err := Database.Query("select id from student where name = ?", st.Name)
	defer rows.Close()
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		rows.Scan(&id)
	}

	st.ID = id
	return
}
