package main

import (
	"database/sql"
)

const (
	NewSub             int = 0
	SubBeingLookedAt   int = 1
	SubGradedByStudent int = 2
	SubGradedByTeacher int = 3
)

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
