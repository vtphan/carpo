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

func flagSubmissionHandler(w http.ResponseWriter, r *http.Request) {

	teacher_id := 0

	if user_id := r.Context().Value("user_id"); user_id != nil {
		teacher_id = user_id.(int)
	}

	fmt.Printf("TeacherID: %v\n", teacher_id)

	subs := make([]FlagSubmission, 0)

	switch r.Method {
	case http.MethodGet:

		log.Printf("Fetching all flagged submission...\n")
		// Only Submissions 2 (not snapshot 1)
		// Only Flagged Submisions/Not Unflagged
		// Only Ungraded (status = 0 )
		sql := "select f.id, f.submission_id, f.problem_id, f.student_id, f.teacher_id, subs.code, s.name, f.created_at, f.updated_at from flagged as f inner join submission as subs on f.submission_id = subs.id INNER join  student as s on f.student_id = s.id where f.soft_delete = 0 and subs.snapshot=2 and subs.status=0"

		s := FlagSubmission{}
		rows, err := Database.Query(sql)
		defer rows.Close()
		if err != nil {
			log.Printf("Error querying db flagSubmissionHandler GET. Err: %v", err)
			return
		}

		for rows.Next() {
			rows.Scan(&s.ID, &s.SubmissionID, &s.ProblemID, &s.StudentID, &s.TeacherID, &s.Code, &s.StudentName, &s.CreatedAt, &s.UpdatedAt)
			subs = append(subs, s)
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
		if pid == 0 || sub_id == 0 || student_id == 0 {
			log.Printf("Failed to save flagged submission to DB. Err. %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Invalid request body.",
				http.StatusInternalServerError)
			return
		}
		flag_id := 0

		sqlStmt := "select id from flagged where submission_id = ?"
		err = Database.QueryRow(sqlStmt, sub_id).Scan(&flag_id)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = AddFlaggedSubmissionSQL.Exec(sub_id, pid, student_id, teacher_id, time.Now(), time.Now())
				if err != nil {
					log.Printf("Failed to save flagged submission to DB. Err. %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save flagged submission to DB.",
						http.StatusInternalServerError)
					return
				}

			}
			log.Printf("Failed to query flagged submission to DB. Err. %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to query flagged submission to DB.",
				http.StatusInternalServerError)
			return

		}
		if flag_id != 0 {
			// Update the row
			Database.Exec("Update flagged set soft_delete=0 where id = ?", flag_id)
		}
		w.WriteHeader(http.StatusCreated)
		resp := []byte(`{"msg": "Submission flagged successfully."}`)
		fmt.Fprint(w, string(resp))

	case http.MethodDelete:
		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}

		fmt.Printf("Body: %v", body)
		fid, _ := strconv.Atoi(fmt.Sprintf("%v", body["flag_id"]))
		if fid == 0 {
			log.Printf("Failed to soft-delete flagged submission. Empty Flag ID")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to soft-delete flagged submission.",
				http.StatusInternalServerError)
			return
		}

		stmt, err := Database.Prepare("UPDATE flagged SET soft_delete=?, updated_at=? where id=?")
		if err != nil {
			log.Printf("SQL Error %v. Err: %v", stmt, err)
		}
		_, err = stmt.Exec(1, time.Now(), fid)
		if err != nil {
			log.Printf("Failed to soft-delete flagged submission in DB. Err. %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to soft-delete flagged submission in DB.",
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		resp := []byte(`{"msg": "Submission Unflagged successfully."}`)
		fmt.Fprint(w, string(resp))

		// case http.MethodOptions:
		// 	w.Header().Set("Access-Control-Allow-Origin", "*")
		// 	w.Header().Set("Access-Control-Allow-Methods", "*")
		// 	w.Header().Set("Access-Control-Allow-Headers", "*")
		// 	w.Header().Set("Access-Control-Max-Age", "86400")
		// 	w.WriteHeader(http.StatusOK)
		// 	return

	}
	// }

}
