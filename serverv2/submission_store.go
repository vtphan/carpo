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
	GetUserNameByID(int) (User, error)
	IsExpired(int) (bool, error)
	GetOnWatchSnapshots() ([]FlagSubmission, error)
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
