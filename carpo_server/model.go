package main

import (
	"time"
)

type Configuration struct {
	IP       string
	Port     int
	Database string
	Address  string
}

type Submission struct {
	ID        int       `json:"id" db:"id"`
	ProblemID int       `json:"problem_id" db:"problem_id"`
	Info      string    `json:"info"`
	Message   string    `json:"message" db:"message"`
	Code      string    `json:"code" db:"code"`
	StudentID int       `json:"student_id" db:"student_id"`
	Name      string    `json:"student_name" db:"name"`
	Status    int       `json:"status" db:"status"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Grade struct {
	ID           int    `json:"id"`
	TeacherID    int    `json:"teacher_id"`
	SubmissionID int    `json:"submission_id"`
	StudnetID    int    `json:"student_id"`
	ProblemID    int    `json:"problem_id"`
	Score        int    `json:"score"`
	Code         string `json:"code"`
	Comment      string `json:"comment"`
}

type Feedback struct {
	ID            int    `json:"id"`
	ProblemID     int    `json:"problem_id"`
	Message       string `json:"message"`
	CodeFeedback  string `json:"code_feedback"`
	Comment       string `json:"comment"`
	LastUpdatedAt string `json:"last_updated_at"`
}

type StudentSubmissionStatus struct {
	ProblemID    int
	SubmissionID int
	Submitted    string
	Score        int
	GradeAt      string
	HasFeedback  int
	FeedbackAt   string
}

type ProblemGradeStatus struct {
	ProblemID        int
	Ungraded         int
	Correct          int
	Incorrect        int
	ProblemStatus    int
	PublishedDate    time.Time
	UnpublishedDated time.Time
}
type Response struct {
	Data      []map[string]interface{} `json:"data"`
	Remaining int
}
