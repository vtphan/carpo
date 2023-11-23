package main

import (
	"fmt"
	"strings"
	"time"
)

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Mode      int       `json:"mode"`
	Status    int       `json:"status,omitempty"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TagProblem struct {
	TagID     int       `json:"tag_id"`
	ProblemID int       `json:"problem_id"`
	Notes     string    `json:"notes"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TagSubmission struct {
	TagID        int       `json:"tag_id"`
	SubmissionID int       `json:"submission_id"`
	Notes        string    `json:"notes"`
	UserID       int       `json:"user_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type TagList struct {
	TagID int    `json:"tag_id"`
	Name  string `json:"name"`
	Mode  int    `json:"mode"`
	Count int    `json:"count"`
}

type TagStore interface {
	GetTags(int) ([]Tag, error)
	CreateTag(Tag) (int, error)
	DeleteTag(int) error
	SaveProblemTag(TagProblem) error
	DeleteProblemTag(int, int) error
	SaveSubmissionTag(TagSubmission) error
	DeleteSubmissionTag(int, int) error
	GetAllTags() ([]TagList, error)
}

func (db *Database) GetTags(mode int) ([]Tag, error) {
	tags := make([]Tag, 0)
	rows, err := db.DB.Query("SELECT id, name, mode, user_id, created_at, updated_at from tags where status = 1 order by created_at asc")
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Tag{}
		rows.Scan(&t.ID, &t.Name, &t.Mode, &t.UserID, &t.CreatedAt, &t.UpdatedAt)

		if t.Mode == mode {
			tags = append(tags, t)
		}
	}
	return tags, err
}

func (db *Database) GetAllTags() ([]TagList, error) {
	idString := []string{"0"}
	tags := make([]TagList, 0)

	// Query from Problem Tags
	rows, err := db.DB.Query("SELECT t1.id, t1.name, t1.mode, count(pt.tag_id) as tag_count from tags as t1 inner join problem_tag as pt on t1.id=pt.tag_id where t1.status = 1 group by (t1.id)")
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		t := TagList{}
		rows.Scan(&t.TagID, &t.Name, &t.Mode, &t.Count)
		tags = append(tags, t)
		idString = append(idString, fmt.Sprint(t.TagID))
	}

	// Query from Submissions Tags
	rows, err = db.DB.Query("SELECT t2.id, t2.name, t2.mode, count(st.tag_id) as tag_count from tags as t2 inner join submission_tag as st on t2.id=st.tag_id where t2.status =1 group by t2.id")
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		t := TagList{}
		rows.Scan(&t.TagID, &t.Name, &t.Mode, &t.Count)
		tags = append(tags, t)
		idString = append(idString, fmt.Sprint(t.TagID))
	}

	// if len(idString) == 0 {
	// 	return tags, err
	// }
	// Newly Created Tags with 0 counts
	fmt.Printf("%+v\n", strings.Join(idString, ","))
	rows, err = db.DB.Query("SELECT id, name, mode from tags where status = 1 and id not in (" + strings.Join(idString, ",") + ")")
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		t := TagList{}
		rows.Scan(&t.TagID, &t.Name, &t.Mode)
		tags = append(tags, t)
	}
	return tags, err

}

func (db *Database) CreateTag(tag Tag) (id int, err error) {
	sqlStatement := `INSERT INTO tags (name, mode, status, user_id, created_at, updated_at) values ( $1, $2, $3, $4, $5, $6) RETURNING id`

	err = db.DB.QueryRow(sqlStatement, tag.Name, tag.Mode, 1, tag.UserID, tag.CreatedAt, tag.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) DeleteTag(tagID int) (err error) {
	sqlStatement := `UPDATE tags SET status = 0 where id = $1`

	_, err = db.DB.Exec(sqlStatement, tagID)

	if err != nil {
		return err
	}
	return
}

func (db *Database) SaveProblemTag(pTag TagProblem) (err error) {
	sqlStatement := `INSERT INTO problem_tag (problem_id, tag_id, notes, user_id, created_at, updated_at) values ( $1, $2, $3, $4, $5, $6)`

	_, err = db.DB.Exec(sqlStatement, pTag.ProblemID, pTag.TagID, pTag.Notes, pTag.UserID, pTag.CreatedAt, pTag.UpdatedAt)

	if err != nil {
		return err
	}

	return err
}

func (db *Database) DeleteProblemTag(tagID int, subID int) (err error) {
	sqlStatement := `DELETE from problem_tag where tag_id = $1 and problem_id = $2;`

	_, err = db.DB.Exec(sqlStatement, tagID, subID)

	if err != nil {
		return err
	}

	return err
}

func (db *Database) SaveSubmissionTag(sTag TagSubmission) (err error) {
	sqlStatement := `INSERT INTO submission_tag (submission_id, tag_id, notes, user_id, created_at, updated_at) values ( $1, $2, $3, $4, $5, $6)`

	_, err = db.DB.Exec(sqlStatement, sTag.SubmissionID, sTag.TagID, sTag.Notes, sTag.UserID, sTag.CreatedAt, sTag.UpdatedAt)

	if err != nil {
		return err
	}

	return err
}

func (db *Database) DeleteSubmissionTag(tagID int, subID int) (err error) {
	sqlStatement := `DELETE from submission_tag where tag_id = $1 and submission_id = $2;`

	_, err = db.DB.Exec(sqlStatement, tagID, subID)

	if err != nil {
		return err
	}

	return err
}
