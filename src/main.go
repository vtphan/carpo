package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "pong") })

	http.HandleFunc("/add_teacher", addUserHandler("teacher"))
	http.HandleFunc("/add_student", addUserHandler("studnet"))

	http.HandleFunc("/students/submissions", studentSubmissionHandler())
	http.HandleFunc("/teachers/submissions", teacherSubmissionHandler())

	http.HandleFunc("/submissions/grade", submissionGradeHandler())

	init_database("my_test_db.sqlite3")

	fmt.Println("serving at port: 8081")

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("Unable to serve gem server at :8081")
	}
}
