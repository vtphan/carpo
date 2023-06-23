package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func deleteProblem(w http.ResponseWriter, r *http.Request) {
	teacher_id := 0

	if user_id := r.Context().Value("user_id"); user_id != nil {
		teacher_id = user_id.(int)
	}

	fmt.Printf("TeacherID: %v\n", teacher_id)

	switch r.Method {
	case http.MethodDelete:
		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		id := int(body["problem_id"].(float64))
		if id != 0 {
			err = archiveProblem(id)
			if err != nil {
				log.Printf("Failed to archive Problem ID: %v. Err: %v", id, err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to archive question to DB.",
					http.StatusInternalServerError)
				return
			}

			d := map[string]interface{}{
				"id":  id,
				"msg": "Question archived successfully.",
			}

			// Delete snapshot as well
			// Remove snapshots from the global map if the problem is expired
			for k := range studentWorkSnapshot {
				expiredProblem := fmt.Sprintf("-%d", id)
				if strings.Contains(k, expiredProblem) {
					fmt.Printf("Deleting student Work Snapshot from map with key: %s.", k)
					delete(studentWorkSnapshot, k)
				}
			}

			data, _ := json.Marshal(d)
			fmt.Fprint(w, string(data))
		} else {
			log.Printf("Invalid Problem ID: %v.\n", body["problem_id"])
		}

	}
}
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
			rows, err := Database.Query("select id, teacher_id, question, format, lifetime from problem where status = 1 order by created_at asc")
			if err != nil {
				log.Printf("Failed to quering problem DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to quering problem DB.",
					http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var (
				id, teacher_id             int
				question, format, lifeTime string
			)

			for rows.Next() {
				rows.Scan(&id, &teacher_id, &question, &format, &lifeTime)

				// Format Expires at:
				ExpiredAt, _ := time.Parse(time.RFC3339, lifeTime)
				question := map[string]interface{}{
					"id":         id,
					"teacher_id": teacher_id,
					"question":   question,
					"format":     format,
					"lifetime":   ExpiredAt.Format("2006-01-02 15:04"),
				}

				// Skip Expired Problem
				if time.Now().After(ExpiredAt) {
					expiredID = append(expiredID, id)

				} else {
					activeQuestions = append(activeQuestions, question)
				}
			}

			// For each downloaded questions, update the StudentProblemStatus Table.
			for _, q := range activeQuestions {
				_, err = Database.Exec("insert into problem (teacher_id, question, format, lifetime, status, created_at, updated_at) values ( ?, ?, ?, ?, ?, ?, ?)", student_id[0], q["id"], 0, time.Now(), time.Now())
				if err != nil {
					log.Printf("Failed to update student problem status(0) to DB. Err. %v\n", err)
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
					log.Printf("Failed to archive Problem ID: %v. Err: %v", id, err)
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
			var questionLife time.Time

			if body["time_limit"] == nil {
				// if no limit is provided,
				// QuestionLife defaults to 90 minutes and status is Active (1)
				questionLife = time.Now().Add((time.Minute * time.Duration(90)))
			} else {
				limit, err := getTimeLimit(fmt.Sprintf("%v", body["time_limit"]))
				if err != nil {
					log.Printf("Failed to parse time_limit of the problem. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, fmt.Sprintf("Failed to parse time_limit of the problem. %v", err),
						http.StatusInternalServerError)
					return

				}
				questionLife = time.Now().Add((time.Minute * time.Duration(limit)))
			}
			res, err := Database.Exec("insert into problem (teacher_id, question, format, lifetime, status, created_at, updated_at) values ( ?, ?, ?, ?, ?, ?, ?)", body["teacher_id"], body["question"], body["format"], questionLife, 1, time.Now(), time.Now())
			// res, err := AddProblemSQL.Exec(body["teacher_id"], body["question"], body["format"], questionLife, 1, time.Now(), time.Now())

			if err != nil {

				log.Printf("Failed to add question to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, fmt.Sprintf("Failed to add question to DB.%v", err),
					http.StatusInternalServerError)
				return

			}

			id, _ := res.LastInsertId()
			log.Printf("Added Problem: %v\n", body)
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
					log.Printf("Failed to archive Problem ID: %v. Err: %v", id, err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to archive question to DB.",
						http.StatusInternalServerError)
					return
				}

				d := map[string]interface{}{
					"id":  id,
					"msg": "Question archived successfully.",
				}

				// Delete snapshot as well
				// Remove snapshots from the global map if the problem is expired
				for k := range studentWorkSnapshot {
					expiredProblem := fmt.Sprintf("-%d", id)
					if strings.Contains(k, expiredProblem) {
						fmt.Printf("Deleting student Work Snapshot from map with key: %s.", k)
						delete(studentWorkSnapshot, k)
					}
				}

				data, _ := json.Marshal(d)
				fmt.Fprint(w, string(data))
			} else {
				log.Printf("Invalid Problem ID: %v.\n", body["problem_id"])
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}

func listProblemsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "15")

		switch r.Method {
		case http.MethodGet:
			rows, err := Database.Query("select id, question, format, lifetime, status from problem")
			if err != nil {
				log.Printf("Failed to list problems. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to list problems.",
					http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var problems []Problem

			for rows.Next() {
				var prob Problem
				if err := rows.Scan(&prob.ID, &prob.Question, &prob.Format, &prob.Lifetime, &prob.Status); err != nil {
					log.Printf("Failed to list problems. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to list problems.",
						http.StatusInternalServerError)
					return
				}
				problems = append(problems, prob)

			}
			data, _ := json.Marshal(problems)
			fmt.Fprint(w, string(data))
		}
	}
}

func archiveProblem(id int) error {

	_, err := Database.Exec("UPDATE problem SET status=?, lifetime=?, updated_at=?  where id=?", 0, time.Now(), time.Now(), id)

	if err != nil {
		log.Printf("SQL Error on archiveProblem. Err: %v", err)
		log.Fatal(err)
	}

	log.Printf("Set Problem status to %v for Problem id: %v.\n", 0, id)

	return err
}

func expireProblems() error {

	rows, err := Database.Query("select id from problem where status = 1  and datetime(lifetime) <= CURRENT_TIMESTAMP order by created_at desc")
	if err != nil {
		return fmt.Errorf("Error querying expired problem. Err: %v", err)
	}
	defer rows.Close()

	var (
		expiredIDs []int
	)

	for rows.Next() {
		var id int
		rows.Scan(&id)
		expiredIDs = append(expiredIDs, id)
	}

	if len(expiredIDs) == 0 {
		return nil
	}

	for _, pid := range expiredIDs {
		err = archiveProblem(pid)
		if err != nil {
			return fmt.Errorf("Failed to auto archive expired Problem ID: %v.. Err: %v\n", pid, err)
		}

		log.Printf("Successfully archived expired Problem ID: %v.\n", pid)

		// Remove snapshots from the global map if the problem is expired
		for k := range studentWorkSnapshot {
			expiredProblem := fmt.Sprintf("-%d", pid)
			if strings.Contains(k, expiredProblem) {
				fmt.Printf("[Cron] Deleting student Work Snapshot from map with key: %s.", k)
				delete(studentWorkSnapshot, k)
			}
		}
	}

	return nil

}

func isExpired(id int) (bool, error) {
	sqlStmt := `SELECT id FROM problem WHERE id = ? and status=0`
	err := Database.QueryRow(sqlStmt, id).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			// a real error happened! you should change your function return
			// to "(bool, error)" and return "false, err" here
			return false, err
		}

		return false, err
	}

	return true, nil

}
