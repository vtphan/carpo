package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func create_tables() {
	execSQL := func(s string) {
		sql_stmt, err := Database.Prepare(s)
		if err != nil {
			log.Fatal(err)
		}
		sql_stmt.Exec()
	}
	execSQL("create table if not exists student (id integer primary key, name text unique)")
	execSQL("create table if not exists teacher (id integer primary key, name text unique)")
	execSQL("create table if not exists problem (id integer primary key, teacher_id integer, question text unique, format string, lifetime timestamp, status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists submission (id integer primary key, problem_id integer, message text, code blob, student_id integer, status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists grade (id integer primary key, teacher_id integer, student_id integer, submission_id integer, score integer, code_feedback blob, comment text, status integer, has_feedback integer default 0, feedback_at timestamp, created_at timestamp, updated_at timestamp, UNIQUE (teacher_id, submission_id))")
	execSQL("create table if not exists student_problem_status (id integer primary key, student_id integer, problem_id integer, tutor_status integer, problem_status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists solution (id integer primary key, problem_id integer unique, code blob, created_at timestamp, updated_at timestamp)")
}

func init_database(db_name string) {
	var err error
	prepare := func(s string) *sql.Stmt {
		stmt, err := Database.Prepare(s)
		if err != nil {
			log.Fatal(err)
		}
		return stmt
	}

	Database, err = sql.Open("sqlite3", db_name)
	if err != nil {
		log.Fatal(err)
	}
	create_tables()
	AddSubmissionSQL = prepare("insert into submission (problem_id, message, code, student_id, status, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?)")
	UpdateSubmissionSQL = prepare("update submission set message=?, code=?, status=?, updated_at=? where problem_id=? and student_id=?")
	AddScoreSQL = prepare("insert into grade (teacher_id, submission_id, student_id, score, status, created_at, updated_at) values (?,?, ?, ?, ?,?,?)")
	AddFeedbackSQL = prepare("insert into grade (teacher_id, submission_id, student_id, score, code_feedback, comment, status, has_feedback, feedback_at, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	UpdateScoreSQL = prepare("update grade set score=?, updated_at=?, status=0 where submission_id=?")
	UpdateFeedbackSQL = prepare("update grade set code_feedback=?, comment=?, has_feedback=1, feedback_at=?, status=0 where teacher_id=? and submission_id=?")
	UpdateScoreFeedbackSQL = prepare("update grade set score=?, code_feedback=?, comment=?, has_feedback=1, feedback_at=?, status=0 where teacher_id=? and submission_id=?")
	UpdateSubmissionFeedbackGivenSQL = prepare("update submission set status=? where id=?")
	AddProblemSQL = prepare("insert into problem (teacher_id, question, format, lifetime, status, created_at, updated_at) values ( ?, ?, ?, ?, ?, ?, ?)")
	AddStudentProblemStatusSQL = prepare("insert into student_problem_status (student_id, problem_id, problem_status, created_at, updated_at) values (?, ?, ?, ?, ?)")
	AddSolutionSQL = prepare("insert into solution (problem_id, code, created_at, updated_at) values (?, ?, ?, ?)")
	UpdateSolutionSQL = prepare("update solution set code=?, updated_at=? where problem_id=?")
}
