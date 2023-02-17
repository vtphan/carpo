package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func teacherSnapshotHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "15")
		// role := "teacher"
		snapshots := make([]Snapshot, 0)

		switch r.Method {
		case http.MethodGet:
			query := r.URL.Query()
			teacher_id, ok := query["id"]
			if !ok || len(teacher_id) < 1 {
				log.Printf("Url Param 'id' is missing.\n")
				http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
				return
			}

			teacher_name, ok := query["name"]
			if !ok || len(teacher_name) < 1 {
				log.Printf("Url Param 'name' is missing.\n")
				http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
				return
			}

			// Get name
			var name string
			rows, err := Database.Query("select name from teacher where id=?", teacher_id[0])
			defer rows.Close()
			if err != nil {
				log.Fatal(err)
			}

			for rows.Next() {
				rows.Scan(&name)
			}

			if name != teacher_name[0] {
				http.Error(w, fmt.Sprintf("You are not authorized to access this status."), http.StatusUnauthorized)
				return
			}

			//
			// for key, value := range studentWorkSnapshot {
			// 	s := Snapshot{}
			// 	combinedKeys := strings.Split(key, "-")
			// 	s.StudentID, _ = strconv.Atoi(combinedKeys[0])
			// 	s.ProblemID, _ = strconv.Atoi(combinedKeys[1])

			// 	student := Studnet{
			// 		ID: s.StudentID,
			// 	}
			// 	s.Name, err = student.GetNameFromID()
			// 	if err != nil {
			// 		log.Printf("Error getting student name from id. Err: %v", err)
			// 		return
			// 	}
			// 	// s.Code = value["code"].(string)
			// 	// s.Time =

			// }

			if len(snapshots) == 0 {
				log.Printf("No new snapshots found.\n")
			}

			resp := Response{}
			sub, _ := json.Marshal(snapshots)

			d := []map[string]interface{}{}
			_ = json.Unmarshal(sub, &d)
			resp.Data = d
			data, _ := json.Marshal(resp)
			fmt.Fprint(w, string(data))

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}
