package main

import (
	"time"
)

type FlagWatchStore interface {
	SaveFlagWatchSubs(FlagSubmission) (int, error)
	UpdateFlagWatchSubs(FlagSubmission) error
	GetFlagSubs(int, int) ([]FlagSubmission, error)
	GetFlagWatchSubID(int, int) (int, error)
	UnFlagWatchSubs(FlagSubmission) error
}

type FlagSubmission struct {
	ID           int       `json:"id"`
	ProblemID    int       `json:"problem_id"`
	SubmissionID int       `json:"submission_id"`
	TeacherID    int       `json:"teacher_id"`
	Mode         int       `json:"mode"`
	Reason       string    `json:"reason"`
	Score        int       `json:"score"`
	StudentID    int       `json:"student_id"`
	StudentName  string    `json:"student_name"`
	Code         string    `json:"code"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (db *Database) GetFlagSubs(isSnapshot int, mode int) ([]FlagSubmission, error) {
	fSubs := make([]FlagSubmission, 0)
	// Only Submissions 2 (not snapshot 1)
	// Only Flagged Submisions/Not Unflagged
	sql := "SELECT fw.id, fw.submission_id, fw.problem_id, subs.user_id, fw.user_id, fw.reason, subs.code, subs.message, u.name, fw.created_at, fw.updated_at, g.score from flag_watch as fw  left join grades as g on fw.submission_id = g.submission_id inner join submissions as subs on fw.submission_id = subs.id INNER join  users as u on  subs.user_id = u.id inner join problems as p on p.id=subs.problem_id where fw.soft_delete = 0 and subs.is_snapshot=$1 and p.status = 1 and fw.mode=$2;"

	rows, err := db.DB.Query(sql, isSnapshot, mode)
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

func (db *Database) GetFlagWatchSubID(subID int, mode int) (id int, err error) {

	sqlStatement := `select id from flag_watch where submission_id = $1 and mode = $2 LIMIT 1;`
	row := db.DB.QueryRow(sqlStatement, subID, mode)
	err = row.Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) SaveFlagWatchSubs(fSub FlagSubmission) (id int, err error) {

	sqlStatement := `INSERT into flag_watch (submission_id, problem_id, user_id, mode, reason, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`

	err = db.DB.QueryRow(sqlStatement, fSub.SubmissionID, fSub.ProblemID, fSub.TeacherID, fSub.Mode, fSub.Reason, fSub.CreatedAt, fSub.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) UpdateFlagWatchSubs(fSub FlagSubmission) (err error) {
	sqlStatement := `UPDATE flag_watch set soft_delete=0, reason=$1 where id = $2`

	_, err = db.DB.Exec(sqlStatement, fSub.Reason, fSub.ID)

	if err != nil {
		return err
	}
	return
}

func (db *Database) UnFlagWatchSubs(fSub FlagSubmission) (err error) {
	sqlStatement := `UPDATE flag_watch set soft_delete=1, updated_at=$1 where id = $2`

	_, err = db.DB.Exec(sqlStatement, time.Now(), fSub.ID)

	if err != nil {
		return err
	}
	return
}
