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
	execSQL("create table if not exists student (id integer primary key, name text unique, uuid BLOB)")
	execSQL("create table if not exists teacher (id integer primary key, name text unique, uuid BLOB)")
	execSQL("create table if not exists problem (id integer primary key, teacher_id integer, question text unique, format string, lifetime timestamp, status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists submission (id integer primary key, problem_id integer, message text, code blob, snapshot integer, student_id integer, status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists grade (id integer primary key, teacher_id integer, student_id integer, submission_id integer, score integer, code_feedback blob, comment text, status integer, has_feedback integer default 0, feedback_at timestamp, created_at timestamp, updated_at timestamp, UNIQUE (teacher_id, submission_id))")
	execSQL("create table if not exists student_problem_status (id integer primary key, student_id integer, problem_id integer, tutor_status integer, problem_status integer, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists solution (id integer primary key, problem_id integer unique, code blob, created_at timestamp, updated_at timestamp)")
	execSQL("create table if not exists flagged (id integer primary key, submission_id integer, problem_id integer, student_id integer, teacher_id integer, soft_delete integer default 0, created_at timestamp, updated_at timestamp, UNIQUE (problem_id, submission_id, student_id))")
	execSQL("create table if not exists watched (id integer primary key, submission_id integer, problem_id integer, student_id integer, teacher_id integer, soft_delete integer default 0, created_at timestamp, updated_at timestamp, UNIQUE (problem_id, submission_id, student_id))")
	// execSQL("alter table solution add column broadcast integer default 0")
	// execSQL("alter table flagged add column reason text")
	// execSQL("alter table watched add column reason text")
}

func init_database(db_name string) {
	var err error
	Database, err = sql.Open("sqlite3", db_name)
	if err != nil {
		log.Fatal(err)
	}
	create_tables()
}
