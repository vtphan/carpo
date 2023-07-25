package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func watchedSubHandler(w http.ResponseWriter, r *http.Request) {

	teacher_id := 0

	if user_id := r.Context().Value("user_id"); user_id != nil {
		teacher_id = user_id.(int)
	}

	// subs := make([]Submission, 0)
	subs := make([]FlagSubmission, 0)

	switch r.Method {
	case http.MethodGet:

		log.Printf("Fetching all watched submission...\n")

		sql := "select f.id, f.submission_id, f.problem_id, f.student_id, f.teacher_id, COALESCE(f.reason,''), subs.code, s.name, f.created_at, f.updated_at from watched as f inner join submission as subs on f.submission_id = subs.id INNER join  student as s on f.student_id = s.id inner join problem as p on p.id=subs.problem_id where f.soft_delete = 0 and subs.snapshot=1 and p.status = 1"

		rows, err := Database.Query(sql)
		if err != nil {
			log.Printf("Error querying db watchedSubHandler GET. Err: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			s := FlagSubmission{}
			rows.Scan(&s.ID, &s.SubmissionID, &s.ProblemID, &s.StudentID, &s.TeacherID, &s.Reason, &s.Code, &s.StudentName, &s.CreatedAt, &s.UpdatedAt)

			// build key to look up on studentWorkSnapshot for latest changes
			key := fmt.Sprintf("%v-%v", s.StudentID, s.ProblemID)
			if val, ok := studentWorkSnapshot[key]; ok {
				subs = append(subs, FlagSubmission{
					ID:           s.ID,
					ProblemID:    s.ProblemID,
					SubmissionID: val.ID,
					StudentID:    s.StudentID,
					TeacherID:    s.TeacherID,
					Reason:       s.Reason,
					StudentName:  s.StudentName,
					Code:         val.Code,
					CreatedAt:    val.CreatedAt,
					UpdatedAt:    val.UpdatedAt,
				})
			} else {
				subs = append(subs, FlagSubmission{
					ID:           s.ID,
					ProblemID:    s.ProblemID,
					SubmissionID: s.SubmissionID,
					StudentID:    s.StudentID,
					TeacherID:    s.TeacherID,
					Reason:       s.Reason,
					StudentName:  s.StudentName,
					Code:         s.Code,
					CreatedAt:    s.CreatedAt,
					UpdatedAt:    s.UpdatedAt,
				})
			}

		}

		submission_id, ok := r.URL.Query()["submission_id"]
		if ok {
			sub_id, _ := strconv.Atoi(submission_id[0])
			if sub_id != 0 {
				for _, item := range subs {
					if item.SubmissionID == sub_id {
						b, _ := json.Marshal(item)
						fmt.Fprint(w, string(b))
						return
					}
				}
				// If request for sub id whose problem is not active
				sql := "select f.id, f.submission_id, f.problem_id, f.student_id, f.teacher_id, COALESCE(f.reason,''), subs.code, s.name, f.created_at, f.updated_at from watched as f inner join submission as subs on f.submission_id = subs.id INNER join  student as s on f.student_id = s.id inner join problem as p on p.id=subs.problem_id where f.soft_delete = 0 and subs.snapshot=1 and subs.id=?"
				rows, err := Database.Query(sql, sub_id)
				if err != nil {
					log.Printf("Error querying db watchedSubHandler GET. Err: %v", err)
					return
				}
				defer rows.Close()

				for rows.Next() {
					s := FlagSubmission{}
					rows.Scan(&s.ID, &s.SubmissionID, &s.ProblemID, &s.StudentID, &s.TeacherID, &s.Reason, &s.Code, &s.StudentName, &s.CreatedAt, &s.UpdatedAt)
					b, _ := json.Marshal(s)
					fmt.Fprint(w, string(b))
					return
				}

			}
		}

		resp := Response{}
		resp.Remaining = len(subs)
		sub, _ := json.Marshal(subs)

		d := []map[string]interface{}{}
		_ = json.Unmarshal(sub, &d)
		resp.Data = d
		data, _ := json.Marshal(resp)
		fmt.Fprint(w, string(data))

	case http.MethodPost:
		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		pid, _ := strconv.Atoi(fmt.Sprintf("%v", body["problem_id"]))
		sub_id, _ := strconv.Atoi(fmt.Sprintf("%v", body["submission_id"]))
		student_id, _ := strconv.Atoi(fmt.Sprintf("%v", body["student_id"]))
		reason := fmt.Sprintf("%v", body["reason"])

		watch_id := 0
		sqlStmt := "select id from watched where submission_id = ?"
		err = Database.QueryRow(sqlStmt, sub_id).Scan(&watch_id)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = Database.Exec("insert or ignore into watched (submission_id, problem_id, student_id, teacher_id, reason, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?)", sub_id, pid, student_id, teacher_id, reason, time.Now(), time.Now())
				if err != nil {
					log.Printf("Failed to watch snapshot to DB. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to watch snapshot to DB.",
						http.StatusInternalServerError)
					return
				}

			} else {
				log.Printf("Failed to query watch submission to DB. Err. %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "Failed to query watch submission to DB.",
					http.StatusInternalServerError)
				return
			}

		}
		if watch_id != 0 {
			// Update the row
			Database.Exec("Update watched set soft_delete=0, reason=? where id = ?", reason, watch_id)
		}

		w.WriteHeader(http.StatusCreated)
		resp := []byte(`{"msg": "Snapshot set on watch successfully."}`)
		fmt.Fprint(w, string(resp))

	case http.MethodDelete:
		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		fid, _ := strconv.Atoi(fmt.Sprintf("%v", body["watch_id"]))
		if fid == 0 {
			log.Printf("Failed to soft-delete watch snapshot. Empty Watch ID")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to soft-delete watch snapshot.",
				http.StatusInternalServerError)
			return
		}

		_, err = Database.Exec("UPDATE watched SET soft_delete=?, updated_at=? where id=?", 1, time.Now(), fid)

		if err != nil {
			log.Printf("Failed to soft-delete watched submission in DB. Err. %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to soft-delete watched submission in DB.",
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		resp := []byte(`{"msg": "Submission Unwatched successfully."}`)
		fmt.Fprint(w, string(resp))

		// case http.MethodOptions:
		// 	w.Header().Set("Access-Control-Allow-Origin", "*")
		// 	w.Header().Set("Access-Control-Allow-Methods", "POST")
		// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// 	w.Header().Set("Access-Control-Max-Age", "3600")
		// 	w.WriteHeader(http.StatusNoContent)
		// 	return

	}

}
