package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "pong") })

	http.HandleFunc("/add_teacher", addUserHandler("teacher"))
	http.HandleFunc("/add_student", addUserHandler("studnet"))

	http.HandleFunc("/problem", problemHandler())

	http.HandleFunc("/students/submissions", studentSubmissionHandler())
	http.HandleFunc("/teachers/submissions", teacherSubmissionHandler())

	http.HandleFunc("/submissions/grade", submissionGradeHandler())

	http.HandleFunc("/teachers/feedbacks", teacherFeedbackHandler())

	http.HandleFunc("/students/get_submission_feedbacks", getSubmissionFeedbacks())

	init_database("my_test_db.sqlite3")

	fmt.Println("serving at port: 8081")

	// Archive expire problems in DB
	ticker := time.NewTicker(10 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf("Running Problem Expiry Checks:\n")
				expireProblems()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("Unable to serve gem server at :8081")
	}
}
