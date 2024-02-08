package main

import (
	"fmt"
	"time"
)

type SolStore interface {
	GetSolution(int) (Solutions, error)
	SaveSolution(Solutions) (int, error)
	UpdateSolution(Solutions) (err error)
	BroadcastSolution(int) error
	IsExpired(int) (bool, error)
}

type Solutions struct {
	ID        int       `json:"id"`
	ProblemID int       `json:"problem_id"`
	UserID    int       `json:"user_id"`
	Code      string    `json:"code"`
	Broadcast int       `json:"broadcast"`
	Format    string    `json:"format"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (db *Database) GetSolution(probID int) (s Solutions, err error) {

	sqlStatement := `SELECT id, problem_id, user_id, code, broadcast, created_at FROM solutions WHERE problem_id=$1 LIMIT 1;`

	row := db.DB.QueryRow(sqlStatement, probID)
	err = row.Scan(&s.ID, &s.ProblemID, &s.UserID, &s.Code, &s.Broadcast, &s.CreatedAt)

	if err != nil {
		return s, err
	}
	return
}

// save solution
func (db *Database) SaveSolution(s Solutions) (id int, err error) {
	fmt.Printf("Req body after parse: %+v\n", s)
	sqlStatement := `INSERT INTO solutions (problem_id, user_id, code, broadcast, created_at, updated_at) values ( $1, $2, $3, $4, $5, $6) RETURNING id`

	err = db.DB.QueryRow(sqlStatement, s.ProblemID, s.UserID, s.Code, s.Broadcast, s.CreatedAt, s.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

// Update solution
func (db *Database) UpdateSolution(s Solutions) (err error) {
	sqlStatement := `UPDATE solutions set code=$1, broadcast=$2, updated_at=$3 where problem_id=$4`

	_, err = db.DB.Exec(sqlStatement, s.Code, s.Broadcast, s.UpdatedAt, s.ProblemID)

	if err != nil {
		return err
	}
	return
}

// Set broadcast to 1
func (db *Database) BroadcastSolution(sID int) (err error) {
	sqlStatement := `UPDATE solutions set broadcast=1 where id=$1`

	_, err = db.DB.Exec(sqlStatement, sID)

	if err != nil {
		return err
	}
	return
}
