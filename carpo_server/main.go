package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func informIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			ip4 := ipnet.IP.To4()
			if ip4 != nil {
				switch {
				case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
				case ip4[0] == 192 && ip4[1] == 168:
				default:
					return ip4.String()
				}
			}
		}
	}
	return ""
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func init_config(filename string) *Configuration {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	config := &Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	if config.IP == "" {
		config.IP = GetLocalIP()
	}
	config.Address = fmt.Sprintf("%s:%d", config.IP, config.Port)
	return config
}

func main() {
	config_file := ""
	flag.StringVar(&config_file, "c", config_file, "json-formatted configuration file.")
	flag.Parse()
	if config_file == "" {
		flag.Usage()
		os.Exit(1)
	}

	Config = init_config(config_file)
	init_database(Config.Database)

	// file, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// log.SetOutput(file)

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "pong") })
	http.HandleFunc("/show_snapshot", showSnapshotHandler())

	http.HandleFunc("/add_teacher", addUserHandler("teacher"))
	http.HandleFunc("/add_student", addUserHandler("studnet"))

	http.HandleFunc("/problem", problemHandler())

	http.Handle("/problems/delete", AuthorizeTeacher(http.HandlerFunc(deleteProblem)))

	http.HandleFunc("/problems/list", listProblemsHandler())

	http.HandleFunc("/students/submissions", studentSubmissionHandler())

	http.Handle("/teachers/submissions", AuthorizeTeacher(http.HandlerFunc(teacherSubmissionHandler)))
	http.Handle("/teachers/snapshots", AuthorizeTeacher(http.HandlerFunc(teacherSnapshotHandler)))

	// http.HandleFunc("/teachers/graded_submissions", gradedSubmissionHandler())

	http.Handle("/submissions/grade", AuthorizeTeacher(http.HandlerFunc(submissionGradeHandler)))
	http.Handle("/submissions/flag", AuthorizeTeacher(http.HandlerFunc(flagSubmissionHandler)))

	http.Handle("/teachers/feedbacks", AuthorizeTeacher(http.HandlerFunc(teacherFeedbackHandler)))

	http.HandleFunc("/students/get_submission_feedbacks", getSubmissionFeedbacks())

	http.Handle("/snapshots/watch", AuthorizeTeacher(http.HandlerFunc(watchedSubHandler)))

	http.HandleFunc("/solution", solutionHandler())
	http.HandleFunc("/students/status", viewStudentSubmissionStatus())

	http.Handle("/problems/status", AuthorizeTeacher(http.HandlerFunc(viewProblemStatus)))
	http.Handle("/solution/broadcast", AuthorizeTeacher(http.HandlerFunc(solutionBroadcast)))

	http.HandleFunc("/problem_detail", problemDetail())

	log.Println("serving at port: 8081")

	// Archive expire problems in DB
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				expireProblems()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	log.Printf("Serving at: %s\n", Config.Address)

	err := http.ListenAndServe(Config.Address, nil)
	if err != nil {
		log.Fatal("Unable to serve carpo server at " + Config.Address)
	}
}
