package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func addUserHandler(role string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var (
			err    error
			msg    string
			exists bool
			req    map[string]interface{}
		)
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		name, ok := req["name"].(string)
		if !ok || name == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			resp := map[string]interface{}{
				"msg": "Empty name is not allowed.",
			}
			j, _ := json.Marshal(resp)
			fmt.Fprint(w, string(j))
			return
		}

		if role == "teacher" {
			s := Teacher{
				Name: name,
			}
			msg, exists, err = s.Add()
			if err != nil {
				log.Printf("Failed to add Teacher. Err: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else {
			// TODO: Student ID should be format U-<INT>
			s := Studnet{
				Name: name,
			}
			msg, exists, err = s.Add()
			if err != nil {
				log.Printf("Failed to add Student. Err: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		if exists {
			w.WriteHeader(http.StatusAlreadyReported)

		} else {
			w.WriteHeader(http.StatusCreated)
		}

		fmt.Fprint(w, string(msg))

	}
}
