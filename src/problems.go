package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func problemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			student_id, ok := r.URL.Query()["student_id"]
			if !ok || len(student_id[0]) < 1 {
				log.Println("Url Param 'student_id' is missing")
				return
			}

			activeQuestions := make([]map[string]interface{}, 0)
			expiredID := make([]int, 0)
			rows, err := Database.Query("select id, teacher_id, question, lifetime from problem where status = 1 order by created_at desc")
			defer rows.Close()
			if err != nil {
				fmt.Errorf("Error quering db. Err: %v", err)
			}

			var (
				id, teacher_id     int
				question, lifeTime string
			)

			for rows.Next() {
				rows.Scan(&id, &teacher_id, &question, &lifeTime)
				question := map[string]interface{}{
					"id":         id,
					"teacher_id": teacher_id,
					"question":   question,
					"lifetime":   lifeTime,
				}

				// Skip Expired Problem
				ExpiredAt, _ := time.Parse(time.RFC3339, lifeTime)
				if time.Now().After(ExpiredAt) {
					expiredID = append(expiredID, id)

				} else {
					activeQuestions = append(activeQuestions, question)
				}
			}

			// For each downloaded questions, update the StudentProblemStatus Table.
			for _, q := range activeQuestions {
				_, err = AddStudentProblemStatusSQL.Exec(student_id[0], q["id"], 0, time.Now(), time.Now())
				if err != nil {
					fmt.Printf("Failed to update student problem status(0) to DB. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to update student problem status(0) to DB.",
						http.StatusInternalServerError)
					return
				}

			}
			// Set the status 0 for expired problems.
			for _, id := range expiredID {
				err = archiveProblem(id)
				if err != nil {
					fmt.Printf("Failed to archive Problem ID: %v. Err: %v", id, err)
				}
			}

			resp := Response{}
			questions, _ := json.Marshal(activeQuestions)
			d := []map[string]interface{}{}
			_ = json.Unmarshal(questions, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))

		case http.MethodPost:
			// QuestionLife defaults to 90 minutes and status is Active (1)
			questionLife := time.Now().Add((time.Minute * time.Duration(90)))
			res, err := AddProblemSQL.Exec(body["teacher_id"], body["question"], questionLife, 1, time.Now(), time.Now())
			if err != nil {

				fmt.Printf("Failed to add question to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to add question to DB.",
					http.StatusInternalServerError)
				return

			}

			id, _ := res.LastInsertId()
			fmt.Printf("Added Problem: %v\n", body)
			w.WriteHeader(http.StatusCreated)
			d := map[string]interface{}{
				"id":  id,
				"msg": "Question saved successfully.",
			}

			data, _ := json.Marshal(d)
			fmt.Fprint(w, string(data))

		case http.MethodDelete:

			id := int(body["problem_id"].(float64))
			if id != 0 {
				err = archiveProblem(id)
				if err != nil {
					fmt.Printf("Failed to archive Problem ID: %v. Err: %v", id, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to archive question to DB.",
						http.StatusInternalServerError)
					return
				}

				d := map[string]interface{}{
					"id":  id,
					"msg": "Question archived successfully.",
				}

				data, _ := json.Marshal(d)
				fmt.Fprint(w, string(data))
			} else {
				fmt.Printf("Invalid Problem ID: %v.\n", body["problem_id"])
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}

func archiveProblem(id int) error {
	stmt, err := Database.Prepare("update problem set status=?, updated_at=?  where id=?")
	if err != nil {
		log.Printf("SQL Error. Err: %v", err)
	}
	fmt.Printf("Set Problem status to %v for Problem id: %v.\n", 0, id)
	_, err = stmt.Exec(0, time.Now(), id)

	return err

}

func expireProblems() error {

	rows, err := Database.Query("select id,lifetime from problem where status = 1  and datetime(lifetime) <= CURRENT_TIMESTAMP order by created_at desc")
	defer rows.Close()
	if err != nil {
		fmt.Errorf("Error quering db. Err: %v", err)
	}

	var (
		expiredIDs []int
	)

	for rows.Next() {
		var id int
		rows.Scan(&id)
		expiredIDs = append(expiredIDs, id)
	}

	if len(expiredIDs) == 0 {
		fmt.Printf("No expired problems in DB.\n")
		return nil
	}

	for _, id := range expiredIDs {
		err = archiveProblem(id)
		if err != nil {
			return fmt.Errorf("Failed to auto archive expired Problem ID: %v.. Err: %v\n", id, err)
		}

		fmt.Printf("Successfully archived expired Problem ID: %v.\n", id)

	}

	return nil

}
