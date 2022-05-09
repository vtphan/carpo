package main

import (
	"log"
)

type Studnet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (st *Studnet) Add() (id int, alreadyExists bool, err error) {

	rows, err := Database.Query("select id from student where name=?", st.Name)
	defer rows.Close()
	if err != nil {
		return 0, false, err
	}

	for rows.Next() {
		rows.Scan(&id)
		log.Printf("User %v already exists as student with ID %v. \n", st.Name, id)
		return id, true, nil
	}

	sqlStatement := `
	insert into student (name) values ($1) returning id`

	err = Database.QueryRow(sqlStatement, st.Name).Scan(&id)
	if err != nil {
		return 0, false, err
	}

	log.Printf("User %v created as student with ID %v.\n", st.Name, id)

	return id, false, nil

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
