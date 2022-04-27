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
		config.IP = informIPAddress()
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

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "pong") })

	http.HandleFunc("/add_teacher", addUserHandler("teacher"))
	http.HandleFunc("/add_student", addUserHandler("studnet"))

	http.HandleFunc("/problem", problemHandler())

	http.HandleFunc("/students/submissions", studentSubmissionHandler())
	http.HandleFunc("/teachers/submissions", teacherSubmissionHandler())

	http.HandleFunc("/submissions/grade", submissionGradeHandler())

	http.HandleFunc("/teachers/feedbacks", teacherFeedbackHandler())

	http.HandleFunc("/students/get_submission_feedbacks", getSubmissionFeedbacks())

	http.HandleFunc("/students/status", viewStudentSubmissionStatus())
	http.HandleFunc("/problems/status", viewProblemStatus())
	http.HandleFunc("/problem_detail", problemDetail())

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

	fmt.Printf("Serving at: %s\n", Config.Address)

	err := http.ListenAndServe(Config.Address, nil)
	if err != nil {
		log.Fatal("Unable to serve carpo server at " + Config.Address)
	}
}
