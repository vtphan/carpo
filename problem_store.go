package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type ProblemStore interface {
	SaveProblem(int, string, string, time.Time) (int, error)
	GetProblems(int) ([]Problem, error)
	ArchiveProblem(int) error
	IsExpired(int) (bool, error)
	ListProblemGradeStatus() ([]ProblemGradeStatus, error)
}

type Problem struct {
	ID       int       `json:"id"`
	Question string    `json:"question"`
	Format   string    `json:"format"`
	Lifetime time.Time `json:"lifetime"`
	Status   int       `json:"status"`
	UserID   int       `json:"user_id"`
}

type ProblemGradeStatus struct {
	ProblemID       int       `json:"problem_id" db:"problem_id"`
	Question        string    `json:"question"`
	SolutionID      int       `json:"solution_id"`
	Solution        string    `json:"solution_code"`
	Ungraded        int       `json:"ungraded"`
	Correct         int       `json:"correct"`
	Incorrect       int       `json:"incorrect"`
	OnWatch         int       `json:"on_watch"`
	ProblemStatus   int       `json:"status"`
	PublishedDate   time.Time `json:"published_at"`
	UnpublishedDate time.Time
	LifeTime        time.Time `json:"lifetime"`
	ExpiresAt       string
	Tag             []Tag `json:"tag,omitempty"`
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
func (db *Database) GetProblems(StudentID int) ([]Problem, error) {
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

	// Set Student Checkout Status
	for _, problem := range activeQuestions {
		sqlStatement := `INSERT into student_checkout_status (user_id, problem_id, created_at, updated_at) values ($1, $2, $3, $4);`

		_, err = db.DB.Exec(sqlStatement, StudentID, problem.ID, time.Now(), time.Now())

		if err != nil {
			return activeQuestions, err
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

// problem status page
func (db *Database) ListProblemGradeStatus() ([]ProblemGradeStatus, error) {
	problemOnWatch := map[int]int{}
	ids := []int{}

	// Get Problem Grading Status
	pGradeStats := make([]ProblemGradeStatus, 0)

	// Get On Watch for the Problem
	rows, err := db.DB.Query(fmt.Sprintf("select problem_id, count(*) as on_watch from flag_watch where mode=1 group by problem_id "))

	if err != nil {
		log.Printf("Error quering db ListProblemGradeStatus for On Watch. Err: %v", err)
		return pGradeStats, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			problem_id, on_watch int64
		)
		rows.Scan(&problem_id, &on_watch)
		problemOnWatch[int(problem_id)] = int(on_watch)

	}

	rows, err = db.DB.Query("select p.id as problem_id, p.question, p.created_at, p.lifetime, p.status, sum(case when s.status in (0,1) and s.is_snapshot=2 then 1 end) as ungraded, sum(case when g.score = 1 then 1 end) as correct, sum(case when g.score = 2 then 1 end) as incorrect, sol.id as solution_id, sol.code as solution_code from problems as p left join submissions as s on p.id = s.problem_id left join grades as g on s.id = g.submission_id LEFT join solutions as sol on sol.problem_id= p.id group by p.id, sol.id, sol.code order by p.id asc;")
	if err != nil {
		log.Printf("Error quering db ListProblemGradeStatus. Err: %v", err)
		return pGradeStats, err
	}
	defer rows.Close()

	for rows.Next() {
		pGradeStat := problemStatus(rows, problemOnWatch)
		// pGradeStat.ExpiresAt = fmt.Sprintf("To be due in %s", fmtDuration(pGradeStat.LifeTime.Sub(time.Now())))
		pGradeStats = append(pGradeStats, pGradeStat)
		ids = append(ids, pGradeStat.ProblemID)
	}

	// convert id from []int to []string
	stringIDs := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	// Get Tags associated with the submissions.
	sql := "select pt.tag_id, pt.problem_id, t.name from problem_tag as pt inner join tags as t on pt.tag_id = t.id where pt.problem_id in (" + stringIDs + ")"
	fmt.Printf("Sql: ", sql)

	rows, err = db.DB.Query(sql)
	if err != nil {
		return pGradeStats, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Tag{}
		pID := 0
		rows.Scan(&t.ID, &pID, &t.Name)
		for idx, prob := range pGradeStats {
			if prob.ProblemID == pID {
				pGradeStats[idx].Tag = append(pGradeStats[idx].Tag, t)
			}
		}
	}

	return pGradeStats, err
}

func problemStatus(rows *sql.Rows, problemOnWatch map[int]int) (pGradeStat ProblemGradeStatus) {

	var (
		ungraded, correct, incorrect sql.NullInt64
	)

	rows.Scan(&pGradeStat.ProblemID, &pGradeStat.Question, &pGradeStat.PublishedDate, &pGradeStat.LifeTime, &pGradeStat.ProblemStatus, &ungraded, &correct, &incorrect, &pGradeStat.SolutionID, &pGradeStat.Solution)

	if !ungraded.Valid {
		ungraded.Int64 = 0
	}
	pGradeStat.Ungraded = int(ungraded.Int64)

	if !correct.Valid {
		correct.Int64 = 0
	}
	pGradeStat.Correct = int(correct.Int64)

	if !incorrect.Valid {
		incorrect.Int64 = 0
	}
	pGradeStat.Incorrect = int(incorrect.Int64)

	pGradeStat.OnWatch = problemOnWatch[int(pGradeStat.ProblemID)]

	return

}
