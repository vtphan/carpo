package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type SubStore interface {
	SaveSubmission(Submission) (int, error)
	GetSubmissions() ([]Submission, error)
	GetSubmissionByID(int) (Submission, error)
	GetUserNameByID(int) (User, error)
	IsExpired(int) (bool, error)
	GetOnWatchSnapshots() ([]FlagSubmission, error)
	GetSnapTags() ([]TagSnapName, error)
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
	OnWatch   int       `json:"on_watch"`
	WatchID   int       `json:"watch_id"`
	Status    int       `json:"status" db:"status"`
	Tag       []Tag     `json:"tag"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TagSnapName struct {
	TagID        int    `json:"id"`
	TagName      string `json:"name"`
	SubmissionID int    `json:"submission_id"`
	UserID       int    `json:"user_id"`
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
	ids := make([]int, 0)

	// sorting := "lower(users.name) ASC"
	sorting := "submissions.created_at ASC"

	sql := "SELECT submissions.id, message, code, is_snapshot, submissions.user_id, users.name, problem_id, problems.format, submissions.created_at, submissions.updated_at from submissions inner join users on submissions.user_id = users.id and submissions.status = 0 and (submissions.is_snapshot = 2 or submissions.is_snapshot = 3) inner join problems on submissions.problem_id = problems.id where problems.status = 1"

	// combine the sorting option:
	sql = fmt.Sprintf("%s ORDER BY %s", sql, sorting)
	s := Submission{}
	rows, err := db.DB.Query(sql)
	if err != nil {
		return subs, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&s.ID, &s.Message, &s.Code, &s.Snapshot, &s.StudentID, &s.Name, &s.ProblemID, &s.Format, &s.CreatedAt, &s.UpdatedAt)
		subs = append(subs, s)
		ids = append(ids, s.ID)
	}

	if len(subs) == 0 {
		log.Printf("No new submissions found.\n")
		return subs, err
	}

	// convert id from []int to []string
	stringIDs := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	// Get Tags associated with the submissions.
	sql = "select st.tag_id, st.submission_id, t.name from submission_tag as st inner join tags as t on st.tag_id = t.id where st.submission_id in (" + stringIDs + ")"
	// fmt.Printf("Sql: ", sql)

	rows, err = db.DB.Query(sql)
	if err != nil {
		return subs, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Tag{}
		subID := 0
		rows.Scan(&t.ID, &subID, &t.Name)
		for idx, sub := range subs {
			if sub.ID == subID {
				subs[idx].Tag = append(subs[idx].Tag, t)
			}
		}
	}

	return subs, err
}

func (db *Database) GetSubmissionByID(submissionID int) (Submission, error) {
	var submission Submission
	
	sql := `SELECT submissions.id, message, code, is_snapshot, submissions.user_id, users.name, problem_id, problems.format, submissions.created_at, submissions.updated_at 
			FROM submissions 
			INNER JOIN users ON submissions.user_id = users.id 
			INNER JOIN problems ON submissions.problem_id = problems.id 
			WHERE submissions.id = $1 AND submissions.status = 0`

	err := db.DB.QueryRow(sql, submissionID).Scan(
		&submission.ID, &submission.Message, &submission.Code, &submission.Snapshot, 
		&submission.StudentID, &submission.Name, &submission.ProblemID, &submission.Format, 
		&submission.CreatedAt, &submission.UpdatedAt)

	if err != nil {
		return submission, err
	}

	// Get associated tags
	tagSQL := `SELECT st.tag_id, t.name FROM submission_tag as st 
			   INNER JOIN tags as t ON st.tag_id = t.id 
			   WHERE st.submission_id = $1`
	
	rows, err := db.DB.Query(tagSQL, submissionID)
	if err != nil {
		return submission, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return submission, err
		}
		submission.Tag = append(submission.Tag, tag)
	}

	return submission, nil
}

func (db *Database) GetSnapTags() ([]TagSnapName, error) {
	tags := make([]TagSnapName, 0)
	sql := "select t.id, t.name, sub.id, sub.user_id from submission_tag as sub_tag inner join submissions as sub on sub_tag.submission_id = sub.id inner join tags as t on sub_tag.tag_id = t.id  inner join problems as p on p.id = sub.problem_id where sub.is_snapshot=1 and p.status=1;"

	rows, err := db.DB.Query(sql)
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		t := TagSnapName{}
		rows.Scan(&t.TagID, &t.TagName, &t.SubmissionID, &t.UserID)
		tags = append(tags, t)
	}

	return tags, err
}

func (db *Database) GetOnWatchSnapshots() ([]FlagSubmission, error) {
	fSubs := make([]FlagSubmission, 0)
	sql := "SELECT fw.id, fw.submission_id, fw.problem_id, subs.user_id, fw.user_id, fw.reason, subs.code, subs.message, u.name, fw.created_at, fw.updated_at, g.score from flag_watch as fw  left join grades as g on fw.submission_id = g.submission_id inner join submissions as subs on fw.submission_id = subs.id INNER join  users as u on  subs.user_id = u.id inner join problems as p on p.id=subs.problem_id where fw.soft_delete = 0 and p.status = 1;"

	rows, err := db.DB.Query(sql)
	if err != nil {
		return fSubs, err
	}
	defer rows.Close()

	for rows.Next() {
		fsub := FlagSubmission{}
		rows.Scan(&fsub.ID, &fsub.SubmissionID, &fsub.ProblemID, &fsub.StudentID, &fsub.TeacherID, &fsub.Reason, &fsub.Code, &fsub.Message, &fsub.StudentName, &fsub.CreatedAt, &fsub.UpdatedAt, &fsub.Score)
		fSubs = append(fSubs, fsub)
	}

	return fSubs, err
}
