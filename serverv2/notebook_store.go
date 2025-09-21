package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotebookStore interface {
	SaveNotebook(title string, mode int, path string, availableTo *time.Time, endTime *time.Time, userID int) (string, error)
	GetNotebooks() ([]AssignmentNotebook, error)
	GetNotebookByUUID(notebookUUID string) (AssignmentNotebook, error)
	GetAvailableNotebooks() ([]AssignmentNotebook, error)
	SaveStudentSubmission(notebookID int, title string, path string, submissionStatus int, fileCreatedAt *time.Time, userID int) error
}

type AssignmentNotebook struct {
	ID            int        `json:"id"`
	NotebookUUID  string     `json:"notebook_uuid"`
	Title         string     `json:"title"`
	Mode          int        `json:"mode"`
	Path          string     `json:"path"`
	AvailableTill *time.Time `json:"available_till"`
	EndTime       *time.Time `json:"end_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (db *Database) SaveNotebook(title string, mode int, path string, availableTo *time.Time, endTime *time.Time, userID int) (string, error) {
	notebookUUID := uuid.New().String()

	sqlStatement := `
		INSERT INTO assignment_notebooks (notebook_uuid, title, mode, path, available_till, end_time, user_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING notebook_uuid;`

	now := time.Now()
	var returnedUUID string

	err := db.DB.QueryRow(sqlStatement, notebookUUID, title, mode, path, availableTo, endTime, userID, now, now).Scan(&returnedUUID)

	if err != nil {
		return "", fmt.Errorf("failed to save notebook: %v", err)
	}

	return returnedUUID, nil
}

func (db *Database) GetNotebooks() ([]AssignmentNotebook, error) {
	sqlStatement := `
		SELECT id, notebook_uuid, title, mode, path, available_till, end_time, created_at, updated_at
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
			&notebook.AvailableTill,
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
		SELECT id, notebook_uuid, title, mode, path, available_to, end_time, created_at, updated_at
		FROM assignment_notebooks 
		WHERE notebook_uuid = $1;`

	var notebook AssignmentNotebook
	err := db.DB.QueryRow(sqlStatement, notebookUUID).Scan(
		&notebook.ID,
		&notebook.NotebookUUID,
		&notebook.Title,
		&notebook.Mode,
		&notebook.Path,
		&notebook.AvailableTill,
		&notebook.EndTime,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
	)

	if err != nil {
		return AssignmentNotebook{}, fmt.Errorf("failed to get notebook: %v", err)
	}

	return notebook, nil
}

func (db *Database) GetAvailableNotebooks() ([]AssignmentNotebook, error) {
	sqlStatement := `
		SELECT id, notebook_uuid, title, mode, path, available_till, end_time, created_at, updated_at
		FROM assignment_notebooks
		WHERE available_till >= NOW()
		ORDER BY created_at DESC;`

	rows, err := db.DB.Query(sqlStatement)
	if err != nil {
		return nil, fmt.Errorf("failed to get available notebooks: %v", err)
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
			&notebook.AvailableTill,
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
