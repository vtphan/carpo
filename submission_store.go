package main

import (
	"fmt"
	"log"
	"time"
)

type SubStore interface {
	SaveSubmission(Submission) (int, error)
	GetSubmissions() ([]Submission, error)
	GetUserNameByID(int) (User, error)
	IsExpired(int) (bool, error)
}

type Submission struct {
	ID        int       `json:"id" db:"id"`
	ProblemID int       `json:"problem_id" db:"problem_id"`
	Format    string    `json:"format"`
	StudentID int       `json:"student_id" db:"user_id"`
	Name      string    `json:"student_name" db:"name"`
	Message   string    `json:"message" db:"message"`
	Code      string    `json:"code" db:"code"`
	Snapshot  int       `json:"snapshot" db:"is_snapshot"`
	Status    int       `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (s *Submission) IsAllowed() bool {
	// Check if preSubTime is greater than 30 sec.
	if prevSubTime, ok := studentLastSubmission[s.StudentID]; ok {
		return time.Now().Sub(prevSubTime).Seconds() >= 30.0
	}
	// if not found, the submission is new
	return true
}

func (db *Database) SaveSubmission(s Submission) (id int, err error) {

	sqlStatement := `INSERT into submissions (problem_id, user_id, message, code, is_snapshot, status, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`

	err = db.DB.QueryRow(sqlStatement, s.ProblemID, s.StudentID, s.Message, s.Code, s.Snapshot, s.Status, s.CreatedAt, s.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) GetSubCodeFromID(subID int) (sCode string, err error) {
	rows, err := db.DB.Query("SELECT code from submissions where id = $1", subID)
	if err != nil {

		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&sCode)
	}
	return
}

// TODO: Sorting by Name and creation time
func (db *Database) GetSubmissions() ([]Submission, error) {
	subs := make([]Submission, 0)

	// sorting := "lower(users.name) ASC"
	sorting := "submissions.created_at ASC"

	sql := "SELECT submissions.id, message, code, submissions.user_id, users.name, problem_id, problems.format, submissions.created_at, submissions.updated_at from submissions inner join users on submissions.user_id = users.id and submissions.status = 0 and submissions.is_snapshot = 2 inner join problems on submissions.problem_id = problems.id where problems.status = 1"

	// combine the sorting option:
	sql = fmt.Sprintf("%s ORDER BY %s", sql, sorting)
	s := Submission{}
	rows, err := db.DB.Query(sql)
	if err != nil {
		return subs, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&s.ID, &s.Message, &s.Code, &s.StudentID, &s.Name, &s.ProblemID, &s.Format, &s.CreatedAt, &s.UpdatedAt)
		subs = append(subs, s)
	}

	if len(subs) == 0 {
		log.Printf("No new submissions found.\n")
	}

	return subs, err
}
