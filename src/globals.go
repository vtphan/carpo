package main

import (
	"database/sql"
)

const (
	NewSub             int = 0
	SubBeingLookedAt   int = 1
	SubGradedByStudent int = 2
	SubGradedByTA      int = 3
	SubGradedByTeacher int = 4
)

var GradingMessage = map[int]string{
	0: "Ungraded",
	1: "Correct",
	2: "Incorrect",
}

var Database *sql.DB
var AddStudentSQL *sql.Stmt
var AddTeacherSQL *sql.Stmt
var AddSubmissionSQL *sql.Stmt
var UpdateSubmissionSQL *sql.Stmt
var UpdateScoreSQL *sql.Stmt
var AddScoreSQL *sql.Stmt
var AddFeedbackSQL *sql.Stmt
var UpdateFeedbackSQL *sql.Stmt
var UpdateSubmissionFeedbackGivenSQL *sql.Stmt
var AddStudentProblemStatus *sql.Stmt
var UpdateStudentProblemStatus *sql.Stmt
var GetStudentFeedback *sql.Stmt
var UpdateSubmissionStatusSQL *sql.Stmt
var AddProblemSQL *sql.Stmt
var AddStudentProblemStatusSQL *sql.Stmt
