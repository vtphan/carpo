package main

import "fmt"

type Teacher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (teacher *Teacher) Add() (id int, alreadyExists bool, err error) {

	fmt.Printf("Adding Teacher: %v\n", teacher.Name)
	rows, err := Database.Query("select id from teacher where name=?", teacher.Name)
	defer rows.Close()
	if err != nil {
		return 0, false, err
	}

	for rows.Next() {
		rows.Scan(&id)
		fmt.Printf("User %v already exists as teacher with ID %v. \n", teacher.Name, id)
		return id, true, nil
	}

	sqlStatement := `
	insert into teacher (name) values ($1) returning id`

	err = Database.QueryRow(sqlStatement, teacher.Name).Scan(&id)
	if err != nil {
		return 0, false, err
	}

	fmt.Printf("User %v created as teacher with ID %v.\n", teacher.Name, id)

	return id, false, nil

}
