package main

import (
	"fmt"
	"time"
)

type ProblemStore interface {
	SaveProblem(int, string, string, time.Time) (int, error)
	GetProblems() ([]Problem, error)
	ArchiveProblem(int) error
	IsExpired(int) (bool, error)
}

type Problem struct {
	ID       int       `json:"id"`
	Question string    `json:"question"`
	Format   string    `json:"format"`
	Lifetime time.Time `json:"lifetime"`
	Status   int       `json:"status"`
	UserID   int       `json:"user_id"`
}

func (db *Database) IsExpired(id int) (bool, error) {
	var status int
	sqlStmt := `SELECT id, status FROM problems WHERE id = $1;`
	err := db.DB.QueryRow(sqlStmt, id).Scan(&id, &status)
	if err != nil {
		return false, err
	}

	return status == 0, nil
}

// save problem
func (db *Database) SaveProblem(user_id int, question string, format string, lifetime time.Time) (id int, err error) {
	sqlStatement := `INSERT INTO problems (user_id, question, format, lifetime, status, created_at, updated_at) values ( $1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = db.DB.QueryRow(sqlStatement, user_id, question, format, lifetime, 1, time.Now(), time.Now()).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

// get active problem
func (db *Database) GetProblems() ([]Problem, error) {
	activeQuestions := make([]Problem, 0)
	expiredID := make([]int, 0)
	rows, err := db.DB.Query("SELECT id, user_id, question, format, lifetime from problems where status = 1 order by created_at asc")
	if err != nil {
		return activeQuestions, err
	}
	defer rows.Close()

	var (
		id, teacher_id             int
		question, format, lifeTime string
	)

	for rows.Next() {
		rows.Scan(&id, &teacher_id, &question, &format, &lifeTime)

		// Format Expires at:
		ExpiredAt, _ := time.Parse(time.RFC3339, lifeTime)
		question := Problem{
			ID:       id,
			UserID:   teacher_id,
			Question: question,
			Format:   format,
			Lifetime: ExpiredAt,
		}

		// Skip Expired Problem
		if time.Now().After(ExpiredAt) {
			expiredID = append(expiredID, id)
		} else {
			activeQuestions = append(activeQuestions, question)
		}
	}
	return activeQuestions, err
}

// archive inactive problems via cron
func (db *Database) ArchiveProblems() ([]int, error) {

	expiredIDs := make([]int, 0)
	rows, err := db.DB.Query("select id from problem where status = 1  and lifetime <= CURRENT_TIMESTAMP order by created_at desc")
	if err != nil {
		return expiredIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		rows.Scan(&id)
		expiredIDs = append(expiredIDs, id)
	}

	if len(expiredIDs) == 0 {
		return expiredIDs, err
	}

	fmt.Printf("Expired: %v\n", expiredIDs)

	for _, pid := range expiredIDs {
		err = db.ArchiveProblem(pid)
		if err != nil {
			return expiredIDs, err
		}
	}

	return expiredIDs, err
}

// archive problem
// archive inactive problems via cron
func (db *Database) ArchiveProblem(id int) error {

	_, err := db.DB.Exec("UPDATE problems SET status=$1, lifetime=$2, updated_at=$3  where id=$4", 0, time.Now(), time.Now(), id)

	if err != nil {
		return err
	}

	return err
}
