package main

import (
	"log"

	"github.com/google/uuid"
)

type Studnet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (st *Studnet) Add() (id int, uid string, alreadyExists bool, err error) {

	log.Printf("Adding Student: %v\n", st.Name)
	rows, err := Database.Query("select id, uuid from student where name=?", st.Name)
	if err != nil {
		return 0, "", false, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&id, &uid)
		log.Printf("User %v already exists as student with ID %v. \n", st.Name, id)
		// return id, uid, true, nil
	}

	// Update already created uid
	if id != 0 {
		uid = uuid.New().String()
		log.Printf("Updating the UID for the student")
		Database.Exec("UPDATE student SET uuid = ? WHERE id = ?", uid, id)
		return id, uid, true, nil
	}

	uid = uuid.New().String()

	sqlStatement := `
	insert into student (name, uuid) values ($1, $2) returning id`

	err = Database.QueryRow(sqlStatement, st.Name, uid).Scan(&id)
	if err != nil {
		return 0, uid, false, err
	}

	log.Printf("User %v created as student with ID %v.\n", st.Name, id)

	return id, uid, false, nil

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
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&id)
	}

	st.ID = id
	return
}

func (st *Studnet) GetNameFromID() (name string, err error) {

	rows, err := Database.Query("select name from student where id = ?", st.ID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&name)
	}

	st.Name = name
	return
}
