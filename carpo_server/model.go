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
	Format    string    `json:"format"`
	Info      string    `json:"info"`
	Message   string    `json:"message" db:"message"`
	Code      string    `json:"code" db:"code"`
	Snapshot  int       `json:"snapshot" db:"snapshot"`
	StudentID int       `json:"student_id" db:"student_id"`
	Name      string    `json:"student_name" db:"name"`
	Status    int       `json:"status" db:"status"`
	Score     int       `json:"score"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Snapshot struct {
	ID        int       `json:"id" db:"id"`
	ProblemID int       `json:"problem_id" db:"problem_id"`
	Code      string    `json:"code" db:"code"`
	StudentID int       `json:"student_id" db:"student_id"`
	Name      string    `json:"student_name" db:"name"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type GradedSubmission struct {
	ID              int       `json:"id"`
	Message         string    `json:"message"`
	Code            string    `json:"code"`
	StudentID       int       `json:"student_id"`
	Score           int       `json:"score"`
	Comment         string    `json:"comment"`
	ProblemID       int       `json:"problem_id"`
	Time            string    `json:"time"`
	ProblemLifeTime time.Time `json:"problem_life_time"`
	SubCreatedAt    time.Time `json:"sub_created_at"`
	GradedCreatedAt time.Time `json:"grade_created_at"`
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
	Snapshot     int
	Code         string
	Submitted    string
	Score        int
	GradeAt      string
	HasFeedback  int
	Feedback     string
	FeedbackAt   string
}
type ProblemStatus struct {
	ProblemID     int
	Question      string
	Status        int
	PublishedDate time.Time
	LifeTime      time.Time
	ExpiresAt     string
	Solution      string
	UploadDate    time.Time
}

type ProblemGradeStatus struct {
	ProblemID       int
	Question        string
	SolutionID      int
	Solution        string
	Ungraded        int
	Correct         int
	Incorrect       int
	ProblemStatus   int
	PublishedDate   time.Time
	UnpublishedDate time.Time
	LifeTime        time.Time
	ExpiresAt       string
}
type Response struct {
	Data      []map[string]interface{} `json:"data"`
	Remaining int
}

type Solutions struct {
	ProblemID int    `json:"problem_id"`
	Solution  string `json:"solution"`
	Format    string `json:"format"`
}

type Problem struct {
	ID       int       `json:"id"`
	Question string    `json:"question"`
	Format   string    `json:"format"`
	Lifetime time.Time `json:"lifetime"`
	Status   int       `json:"status"`
}

type FlagSubmission struct {
	ID           int       `json:"id"`
	ProblemID    int       `json:"problem_id"`
	SubmissionID int       `json:"submission_id"`
	Score        int       `json:"score"`
	StudentID    int       `json:"student_id"`
	TeacherID    int       `json:"teacher_id"`
	Reason       string    `json:"reason"`
	StudentName  string    `json:"student_name"`
	Code         string    `json:"code"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
