package main

import "time"

type Submission struct {
	ID         int       `json:"id" db:"id"`
	QuestionID int       `json:"question_id" db:"question_id"`
	Message    string    `json:"message" db:"message"`
	Code       string    `json:"code" db:"code"`
	StudentID  int       `json:"student_id" db:"student_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Response struct {
	Data []map[string]interface{} `json:"data"`
}
