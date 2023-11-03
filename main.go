package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()

	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:8080"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	psqlInfo := fmt.Sprintf("postgresql://%v:%v@%v/%v?sslmode=disable", os.Getenv("POSTGRES_USR"), os.Getenv("POSTGRES_PWD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DB"))
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	uAPI := UserAPI{&Database{DB: db}}
	pAPI := ProblemAPI{&Database{DB: db}}
	subAPI := SubmissionAPI{&Database{DB: db}}
	gradeAPI := GradeAPI{&Database{DB: db}}

	flagAPI := FlagWatchAPI{&Database{DB: db}}

	solAPI := SolutionAPI{&Database{DB: db}}

	// Register Users
	r.POST("/users", uAPI.RegisterUser)

	// Problems
	r.POST("/problems", pAPI.PublishProblem)
	r.GET("/problems/students/:user_id", pAPI.GetActiveProblems)
	r.DELETE("/problems/:id", pAPI.UnpublishProblem)

	// Solution
	r.POST("/solution", solAPI.SolutionHandler)

	// Submissions & Snapshots
	r.POST("/submissions/students/:user_id", subAPI.SubmissionHandler)

	// Use Middleware for app APIs
	r.Use(appMiddleware(db))
	r.GET("/submissions/teachers", subAPI.GetSubmissionsHandler)
	r.OPTIONS("/submissions/teachers")

	// Grades and Feedbacks
	r.POST("/submissions/grades", gradeAPI.GradeHandler)
	r.OPTIONS("/submissions/grades")

	// Flag Submissions
	r.GET("/submissions/flag", flagAPI.GetFlagSubsHandler)
	r.POST("/submissions/flag", flagAPI.FlagSubHandler)
	r.DELETE("/submissions/flag", flagAPI.DelFlagSubHandler)
	r.OPTIONS("/submissions/flag")

	// Watch Snapshot
	r.GET("/snapshots/teachers", subAPI.GetSnapshotsHandler)
	r.GET("/snapshots/watch", flagAPI.GetWatchSubsHandler)
	r.POST("/snapshots/watch", flagAPI.FlagSubHandler)
	r.DELETE("/snapshots/watch", flagAPI.DelFlagSubHandler)

	r.Run(":8081")
}
