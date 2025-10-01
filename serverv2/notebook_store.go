package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotebookStore interface {
	SaveNotebook(title string, mode int, path string, startTime *time.Time, endTime *time.Time, userID int) (string, error)
	GetNotebooks() ([]AssignmentNotebook, error)
	GetNotebookByUUID(notebookUUID string) (AssignmentNotebook, error)
	UpdateNotebook(int, string, string) error
	DeleteNotebook(int) error
	GetAvailableNotebooks(int) ([]AssignmentNotebook, error)
	GetNotebookEndTime(int) (AssignmentNotebook, error)
	SaveStudentSubmission(notebookID int, title string, path string, submissionStatus int, fileCreatedAt *time.Time, userID int) error
	GetSubmittedNotebooksByID(notebookID int) ([]StudentSubmission, error)
}

type StudentSubmission struct {
	ID               int        `json:"id"`
	NotebookID       int        `json:"notebook_id"`
	Title            string     `json:"title"`
	Path             string     `json:"path"`
	SubmissionStatus int        `json:"submission_status"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	FileCreatedAt    *time.Time `json:"file_created_at"`
	UserID           int        `json:"user_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AssignmentNotebook struct {
	ID           int        `json:"id"`
	NotebookUUID string     `json:"notebook_uuid"`
	Title        string     `json:"title"`
	Mode         int        `json:"mode"`
	Path         string     `json:"path"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (db *Database) SaveNotebook(title string, mode int, path string, startTime *time.Time, endTime *time.Time, userID int) (string, error) {
	notebookUUID := uuid.New().String()

	sqlStatement := `
		INSERT INTO assignment_notebooks (notebook_uuid, title, mode, path, start_time, end_time, user_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING notebook_uuid;`

	now := time.Now()
	var returnedUUID string

	err := db.DB.QueryRow(sqlStatement, notebookUUID, title, mode, path, startTime, endTime, userID, now, now).Scan(&returnedUUID)

	if err != nil {
		return "", fmt.Errorf("failed to save notebook: %v", err)
	}

	return returnedUUID, nil
}

func (db *Database) GetNotebooks() ([]AssignmentNotebook, error) {
	sqlStatement := `
		SELECT id, notebook_uuid, title, mode, path, start_time, end_time, created_at, updated_at
		FROM assignment_notebooks
		ORDER BY created_at DESC;`

	rows, err := db.DB.Query(sqlStatement)
	if err != nil {
		return nil, fmt.Errorf("failed to get notebooks: %v", err)
	}
	defer rows.Close()

	var notebooks []AssignmentNotebook
	for rows.Next() {
		var notebook AssignmentNotebook
		err := rows.Scan(
			&notebook.ID,
			&notebook.NotebookUUID,
			&notebook.Title,
			&notebook.Mode,
			&notebook.Path,
			&notebook.StartTime,
			&notebook.EndTime,
			&notebook.CreatedAt,
			&notebook.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notebook: %v", err)
		}
		notebooks = append(notebooks, notebook)
	}

	return notebooks, nil
}

func (db *Database) GetNotebookByUUID(notebookUUID string) (AssignmentNotebook, error) {
	sqlStatement := `
		SELECT id, notebook_uuid, title, mode, path, start_time, end_time, created_at, updated_at
		FROM assignment_notebooks 
		WHERE notebook_uuid = $1;`

	var notebook AssignmentNotebook
	err := db.DB.QueryRow(sqlStatement, notebookUUID).Scan(
		&notebook.ID,
		&notebook.NotebookUUID,
		&notebook.Title,
		&notebook.Mode,
		&notebook.Path,
		&notebook.StartTime,
		&notebook.EndTime,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
	)

	if err != nil {
		return AssignmentNotebook{}, fmt.Errorf("failed to get notebook: %v", err)
	}

	return notebook, nil
}

func (db *Database) GetAvailableNotebooks(mode int) ([]AssignmentNotebook, error) {
	sqlStatement := `
		SELECT id, notebook_uuid, title, mode, path, start_time, end_time, created_at, updated_at
		FROM assignment_notebooks
		WHERE start_time <= NOW() AND end_time >= NOW() AND mode = $1
		ORDER BY created_at DESC;`

	rows, err := db.DB.Query(sqlStatement, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to get available notebooks: %v", err)
	}
	defer rows.Close()

	notebooks := make([]AssignmentNotebook, 0)
	for rows.Next() {
		var notebook AssignmentNotebook
		err := rows.Scan(
			&notebook.ID,
			&notebook.NotebookUUID,
			&notebook.Title,
			&notebook.Mode,
			&notebook.Path,
			&notebook.StartTime,
			&notebook.EndTime,
			&notebook.CreatedAt,
			&notebook.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notebook: %v", err)
		}
		notebooks = append(notebooks, notebook)
	}

	return notebooks, nil
}

func (db *Database) SaveStudentSubmission(notebookID int, title string, path string, submissionStatus int, fileCreatedAt *time.Time, userID int) error {
	sqlStatement := `
		INSERT INTO students_notebooks_submissions (notebook_id, title, path, submission_status, submitted_at, file_created_at, user_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	now := time.Now()

	_, err := db.DB.Exec(sqlStatement, notebookID, title, path, submissionStatus, now, fileCreatedAt, userID, now, now)

	if err != nil {
		return fmt.Errorf("failed to save student submission: %v", err)
	}

	return nil
}

func (db *Database) GetNotebookEndTime(notebookID int) (AssignmentNotebook, error) {
	sqlStatement := `
	SELECT id, notebook_uuid, title, mode, path, start_time, end_time, created_at, updated_at
	FROM assignment_notebooks
	WHERE id = $1 Limit 1`

	var notebook AssignmentNotebook

	err := db.DB.QueryRow(sqlStatement, notebookID).Scan(&notebook.ID,
		&notebook.NotebookUUID,
		&notebook.Title,
		&notebook.Mode,
		&notebook.Path,
		&notebook.StartTime,
		&notebook.EndTime,
		&notebook.CreatedAt,
		&notebook.UpdatedAt)
	if err != nil {
		return notebook, fmt.Errorf("failed to get available notebooks: %v", err)
	}

	return notebook, nil
}

func (db *Database) UpdateNotebook(notebookID int, StartTime string, EndTime string) error {
	sqlStatement := `UPDATE assignment_notebooks set start_time=$1, end_time=$2, updated_at=$3 where id=$4;`
	now := time.Now()

	_, err := db.DB.Exec(sqlStatement, StartTime, EndTime, now, notebookID)

	if err != nil {
		return fmt.Errorf("failed to update Notebook: %v", err)
	}

	return nil
}

func (db *Database) DeleteNotebook(notebookID int) error {
	sqlStatement := `DELETE FROM assignment_notebooks where id=$1;`
	_, err := db.DB.Exec(sqlStatement, notebookID)

	if err != nil {
		return fmt.Errorf("failed to delete Notebook: %v", err)
	}

	return nil
}

func (db *Database) GetSubmittedNotebooksByID(notebookID int) ([]StudentSubmission, error) {
	sqlStatement := `
		SELECT DISTINCT ON (user_id) id, notebook_id, title, path, submission_status, submitted_at, file_created_at, user_id, created_at, updated_at
		FROM students_notebooks_submissions 
		WHERE notebook_id = $1 AND submission_status = 3
		ORDER BY user_id, submitted_at DESC;`

	rows, err := db.DB.Query(sqlStatement, notebookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submitted notebooks: %v", err)
	}
	defer rows.Close()

	var submissions []StudentSubmission
	for rows.Next() {
		var submission StudentSubmission
		err := rows.Scan(
			&submission.ID,
			&submission.NotebookID,
			&submission.Title,
			&submission.Path,
			&submission.SubmissionStatus,
			&submission.SubmittedAt,
			&submission.FileCreatedAt,
			&submission.UserID,
			&submission.CreatedAt,
			&submission.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission: %v", err)
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}
