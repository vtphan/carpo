package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type SSEClient struct {
	Channel chan string
	UserID  int
}

type SSEHub struct {
	clients map[int]*SSEClient
	mutex   sync.RWMutex
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[int]*SSEClient),
	}
}

func (h *SSEHub) AddClient(userID int, client *SSEClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.clients[userID] = client
}

func (h *SSEHub) RemoveClient(userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if client, exists := h.clients[userID]; exists {
		close(client.Channel)
		delete(h.clients, userID)
	}
}

func (h *SSEHub) BroadcastToUser(userID int, message string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if client, exists := h.clients[userID]; exists {
		select {
		case client.Channel <- message:
		default:
		}
	}
}

type FeedbackMessage struct {
	EventType    string    `json:"event_type"`
	Grade        int       `json:"grade"`
	SubmissionID int       `json:"submission_id"`
	StudentID    int       `json:"student_id"`
	ProblemID    int       `json:"problem_id"`
	Timestamp    time.Time `json:"timestamp"`
}

var sseHub = NewSSEHub()

func eventsHandler(c *gin.Context) {
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		c.JSON(400, gin.H{"error": "user_id parameter is required"})
		return
	}

	userID := 0
	if _, err := fmt.Sscanf(userIDParam, "%d", &userID); err != nil {
		c.JSON(400, gin.H{"error": "invalid user_id parameter"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	client := &SSEClient{
		Channel: make(chan string, 100),
		UserID:  userID,
	}

	sseHub.AddClient(userID, client)
	defer sseHub.RemoveClient(userID)

	clientGone := c.Request.Context().Done()

	for {
		select {
		case message := <-client.Channel:
			c.SSEvent("message", message)
			c.Writer.Flush()
		case <-clientGone:
			return
		}
	}
}

func main() {

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()

	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://127.0.0.1:8080", "http://localhost:8080", "http://141.225.10.71:8000"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.LoadHTMLGlob("templates/*")

	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", os.Getenv("POSTGRES_USR"), os.Getenv("POSTGRES_PWD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DB"))
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// SSE endpoint for real-time events
	r.GET("/events", eventsHandler)

	uAPI := UserAPI{&Database{DB: db}}
	pAPI := ProblemAPI{&Database{DB: db}}
	subAPI := SubmissionAPI{&Database{DB: db}}
	gradeAPI := GradeAPI{&Database{DB: db}}
	flagAPI := FlagWatchAPI{&Database{DB: db}}
	solAPI := SolutionAPI{&Database{DB: db}}
	tagAPI := TagAPI{&Database{DB: db}}
	agentAPI := AgentAPI{&Database{DB: db}}
	feedbackAPI := FeedbackAPI{&Database{DB: db}}
	feedbackAgentAPI := FeedbackAgentAPI{
		DB: &Database{DB: db},
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: os.Getenv("FEEDBACK_AGENT"),
	}
	notebookAPI := NotebookAPI{&Database{DB: db}}

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

	// Student ask for help
	r.POST("/students/:user_id/ask_for_help", flagAPI.StudentAskForHelp)

	// Student Status page
	r.GET("students/status", viewStudentSubmissionStatus(db))
	r.GET("/solutions/problem/:id", solAPI.GetSolutionByProblemIDHandler)

	// Student Feedback on Problems
	r.GET("/students/:user_id/problems/:problem_id/feedbacks", feedbackAPI.GetFeedbackHandler)
	// Student Rate on feedbacks
	r.PUT("/feedback-ratings", feedbackAgentAPI.UpdateFeedbackRatingHandler)

	// Notebooks
	r.POST("/notebooks", notebookAPI.UploadNotebook)
	r.GET("/notebooks", notebookAPI.GetNotebooks)
	r.GET("/notebooks/:uuid", notebookAPI.GetNotebookByUUID)
	r.GET("/notebooks/students/:user_id/download", notebookAPI.GetAvailableNotebooks)
	r.GET("/notebooks/file", notebookAPI.ServeNotebookFile)
	r.POST("/notebooks/students/:user_id/submit", notebookAPI.SubmitNotebook)
	r.OPTIONS("/notebooks")

	// Use Middleware for app APIs
	r.Use(appMiddleware(db))
	r.GET("/submissions/teachers", subAPI.GetSubmissionsHandler)
	r.GET("/submissions/:id/agent-feedback", feedbackAgentAPI.GetAgentFeedbackByIDHandler)

	r.OPTIONS("/submissions/teachers")

	// Grades and Feedbacks
	r.POST("/submissions/grades", gradeAPI.GradeHandler)
	r.OPTIONS("/submissions/grades")

	// Flag Submissions
	// r.GET("/submissions/flag", flagAPI.GetFlagSubsHandler)
	// r.POST("/submissions/flag", flagAPI.FlagSubHandler)
	// r.DELETE("/submissions/flag", flagAPI.DelFlagSubHandler)
	// r.OPTIONS("/submissions/flag")

	// Watch Snapshot
	r.GET("/snapshots/teachers", subAPI.GetSnapshotsHandler)
	r.GET("/snapshots/watch", flagAPI.GetWatchSubsHandler)
	r.POST("/snapshots/watch", flagAPI.FlagSubHandler)
	r.DELETE("/snapshots/watch", flagAPI.DelFlagSubHandler)

	// Tag
	r.GET("/tags", tagAPI.GetTagHandler)
	r.POST("/tags", tagAPI.SaveTagHandler)
	r.POST("/tags/:id", tagAPI.UpdateTagHandler)
	r.OPTIONS("/tags")
	r.DELETE("/tags/:id", tagAPI.DeleteTagHandler)

	// Tag Submissions
	r.POST("/tags/submissions/", tagAPI.TagSubmissionHandler)
	r.DELETE("/tags/:id/submissions/:sid", tagAPI.TagSubmissionDelHandler)

	// Tag Problems
	r.POST("/tags/problems/", tagAPI.TagProblemHandler)
	r.DELETE("/tags/:id/problems/:pid", tagAPI.TagProblemDelHandler)

	r.GET("/tags/tagged", tagAPI.GetAllTagHandler)

	// Agents
	r.GET("/agents", agentAPI.GetAgentHandler)
	r.POST("/agents", agentAPI.SaveAgentHandler)
	r.PUT("/agents/:id", agentAPI.UpdateAgentHandler)
	r.DELETE("/agents/:id", agentAPI.DeleteAgentHandler)
	r.OPTIONS("/agents")

	// Problem Status Page
	r.GET("/problems/status", pAPI.ViewProblemStatus)
	r.OPTIONS("/problems/status")

	// Broadcast solution
	r.PUT("/solutions/:id/broadcast", solAPI.BroadcastSolHandler)
	r.OPTIONS("/solutions")

	r.Run(":8081")
}
