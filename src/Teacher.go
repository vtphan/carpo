package main

import "fmt"

type Teacher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (teacher *Teacher) Add() (msg string, alreadyExists bool, err error) {

	rows, err := Database.Query("select name from teacher where name=?", teacher.Name)
	defer rows.Close()
	if err != nil {
		return "", false, err
	}

	for rows.Next() {
		return fmt.Sprintf("User %v already exists as %v.", teacher.Name, "teacher"), true, nil
	}
	_, err = AddTeacherSQL.Exec(teacher.Name)
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("User %v created as %v.", teacher.Name, "teacher"), false, nil

}
