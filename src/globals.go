package main

import (
	"database/sql"
)

var Database *sql.DB
var AddStudentSQL *sql.Stmt
var AddTeacherSQL *sql.Stmt
var AddSubmissionSQL *sql.Stmt
var AddScoreSQL *sql.Stmt
