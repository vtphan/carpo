package main

import (
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

	id, err = s.SaveSubmission(st.ID)
	if err != nil {
		log.Printf("Failed to Save Submission %v for student %v. Err: %v\n", s, st, err)
	}

	return
}

func (st *Studnet) UpdateSubmission(s Submission) (err error) {

	err = s.UpdateSubmission(st.ID)
	if err != nil {
		log.Printf("Failed to update Submission %v for student %v. Err: %v", s, st, err)
	}
	return
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
