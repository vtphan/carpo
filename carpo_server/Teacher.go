package main

import (
	"log"

	"github.com/google/uuid"
)

type Teacher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (teacher *Teacher) Add() (id int, uid string, alreadyExists bool, err error) {

	log.Printf("Adding Teacher: %v\n", teacher.Name)
	rows, err := Database.Query("select id, uuid from teacher where name=?", teacher.Name)
	defer rows.Close()
	if err != nil {
		return 0, "", false, err
	}

	for rows.Next() {
		rows.Scan(&id, &uid)
		log.Printf("User %v already exists as teacher with ID %v. \n", teacher.Name, id)
		// return id, uid, true, nil
	}

	// Update already created uid
	if id != 0 {
		uid = uuid.New().String()
		log.Printf("Updating the UID for the teacher")
		Database.Exec("UPDATE teacher SET uuid = ? WHERE id = ?", uid, id)
		return id, uid, true, nil
	}

	uid = uuid.New().String()
	sqlStatement := `
	insert into teacher (name, uuid) values ($1, $2) returning id`

	err = Database.QueryRow(sqlStatement, teacher.Name, uid).Scan(&id)
	if err != nil {
		return 0, "", false, err
	}

	log.Printf("User %v created as teacher with ID %v.\n", teacher.Name, id)

	return id, uid, false, nil

}
