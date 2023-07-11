package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
)

func solutionBroadcast(w http.ResponseWriter, r *http.Request) {
	body, err := readRequestBody(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
		return
	}
	sid, _ := strconv.Atoi(fmt.Sprintf("%v", body["solution_id"]))
	if sid == 0 {
		log.Printf("Failed to broadcast solution. Empty Solution ID")
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Failed to broadcast solution.",
			http.StatusInternalServerError)
		return
	}

	_, err = Database.Exec("UPDATE solution SET broadcast=?, updated_at=? where id=?", 1, time.Now(), sid)

	if err != nil {
		log.Printf("Failed to broadcast solution in DB. Err. %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Failed to broadcast solution in DB.",
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := []byte(`{"msg": "Solution broadcasted successfully."}`)
	fmt.Fprint(w, string(resp))

}
func solutionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// Deprecated. Not in Use
		case http.MethodGet:
			// Get all problems
			// Get all solutions
			rows, err := Database.Query("select id, teacher_id, question, format from problem order by created_at asc")
			if err != nil {
				err = fmt.Errorf("Error quering GetSolution problem. Err: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Error reading request body",
					http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var (
				id, teacher_id   int
				question, format string
			)
			Sols := make([]Solutions, 0)

			for rows.Next() {
				rows.Scan(&id, &teacher_id, &question, &format)
				sol := Solutions{
					ProblemID: id,
					Format:    format,
				}

				rows, err := Database.Query("select code from solution where problem_id=?", id)
				if err != nil {
					err = fmt.Errorf("Error quering GetSolution. Err: %v", err)
					return
				}
				defer rows.Close()

				for rows.Next() {
					rows.Scan(&sol.Solution)
				}
				if sol.Solution != "" {
					Sols = append(Sols, sol)
				}

			}

			resp := Response{}
			sols, _ := json.Marshal(Sols)
			d := []map[string]interface{}{}
			_ = json.Unmarshal(sols, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))

		case http.MethodPost:
			problem_id, ok := r.URL.Query()["problem_id"]
			if !ok || len(problem_id[0]) < 1 {
				log.Println("Url Param 'problem_id' is missing")
				http.Error(w, fmt.Sprintf("problem_id is missing."), http.StatusUnprocessableEntity)
				return
			}

			body, err := readRequestBody(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				http.Error(w, "Error reading request body",
					http.StatusInternalServerError)
				return
			}
			code := fmt.Sprintf("%v", body["code"])

			// If the problem is unpublished, the solution should should broadcast.
			broadcast := 0
			pid, _ := strconv.Atoi(problem_id[0])
			expiredProblem, err := isExpired(pid)
			if err != nil {
				log.Printf("Failed to check the expired problem in solutionHandler. Err: %v", err)
				http.Error(w, "Error reading request body",
					http.StatusInternalServerError)
				return
			}
			if expiredProblem {
				broadcast = 1
				log.Printf("Setting solution to broadcast for problem id: %v", pid)
			}

			_, err = Database.Exec("insert into solution (problem_id, code, broadcast, created_at, updated_at) values (?, ?, ?, ?)", problem_id[0], code, broadcast, time.Now(), time.Now())
			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Printf("Solution already exists for problemID %v. Updating...\n", problem_id[0])
					if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
						_, err = Database.Exec("update solution set code=?, broadcast=?, updated_at=? where problem_id=?", code, broadcast, time.Now(), problem_id[0])
						if err != nil {
							log.Printf("Failed to update solution for problemID %+v. Err: %v", problem_id[0], err)
						}
						log.Printf("Solution successfully updated.")
					}

				} else {
					log.Printf("Failed to save Solution for problemID: %v Err. %v\n", problem_id[0], err)
					w.WriteHeader(http.StatusInternalServerError)
					http.Error(w, "Failed to save solution.",
						http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusCreated)
			resp := []byte(`{"msg": "Solution successfully uploaded."}`)
			fmt.Fprint(w, string(resp))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}
