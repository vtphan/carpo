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
	execSQL("create table if not exists submission (id integer primary key, question_id integer, message text, code blob, student_id integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists score (id integer primary key, teacher_id integer, student_id integer, submission_id integer, points integer, created_at timestamp, updated_at timestamp)")

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
	AddStudentSQL = prepare("insert into student (name) values (?)")
	AddTeacherSQL = prepare("insert into teacher (name) values (?)")
	AddSubmissionSQL = prepare("insert into submission (question_id, message, code, student_id, created_at, updated_at) values (?, ?, ?, ?, ?, ?)")
	AddScoreSQL = prepare("insert into score (teacher_id, student_id, submission_id, points, created_at, updated_at) values (?, ?, ?, ?, ?, ?)")

}
