package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type NotebookAPI struct {
	NotebookService NotebookStore
}

var NotebookStatus = map[string]int{
	"downloaded": 1,
	"draft":      2,
	"submitted":  3,
}

var NotebookMode = map[string]int{
	"assignment": 1,
	"exam":       2,
}

const buffer = 10

type UploadNotebookRequest struct {
	Title     string `form:"title" binding:"required"`
	Mode      int    `form:"mode" binding:"required"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	UserID    int    `form:"user_id" binding:"required"`
}

type UpdateNotebookRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (n *NotebookAPI) UploadNotebook(c *gin.Context) {
	var req UploadNotebookRequest

	if err := c.ShouldBind(&req); err != nil {
		log.Infof("Error parsing form data in UploadNotebook. Err: %v", err)
		c.JSON(400, gin.H{"error": "Invalid form data", "details": err.Error()})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("filecontent")
	if err != nil {
		log.Infof("Error getting file from form in UploadNotebook. Err: %v", err)
		c.JSON(400, gin.H{"error": "No file uploaded", "details": err.Error()})
		return
	}

	// Validate file extension
	if filepath.Ext(file.Filename) != ".ipynb" {
		c.JSON(400, gin.H{"error": "File must be a .ipynb notebook file"})
		return
	}

	fileName := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))

	// Parse timestamp strings
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		if parsed, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			startTime = &parsed
		}
	}
	if req.EndTime != "" {
		if parsed, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = &parsed
		}
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads/notebooks"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Errorf("Failed to create uploads directory: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate filename based on title and timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.ipynb", fileName, timestamp)
	// Sanitize filename
	filename = filepath.Base(filename)
	filePath := filepath.Join(uploadsDir, filename)

	// Save the uploaded file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Errorf("Failed to save uploaded file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// Save notebook metadata to database
	notebookUUID, err := n.NotebookService.SaveNotebook(
		req.Title,
		req.Mode,
		filePath,
		startTime,
		endTime,
		req.UserID,
	)
	if err != nil {
		log.Errorf("Failed to save notebook metadata: %v", err)
		// Clean up the file if database save fails
		os.Remove(filePath)
		c.JSON(500, gin.H{"error": "Failed to save notebook metadata", "details": err.Error()})
		return
	}

	response := gin.H{
		"notebook_uuid": notebookUUID,
		"title":         req.Title,
		"mode":          req.Mode,
		"path":          filePath,
		"start_time":    startTime,
		"end_time":      endTime,
		"user_id":       req.UserID,
		"message":       "Notebook uploaded successfully",
	}

	c.JSON(200, response)
}

func (n *NotebookAPI) GetNotebooks(c *gin.Context) {

	notebooks, err := n.NotebookService.GetNotebooks()
	if err != nil {
		log.Errorf("Failed to get notebooks: %v", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve notebooks"})
		return
	}

	c.JSON(200, gin.H{"data": notebooks})
}

func (n *NotebookAPI) GetNotebookByUUID(c *gin.Context) {
	notebookUUID := c.Param("uuid")
	if notebookUUID == "" {
		c.JSON(400, gin.H{"error": "Notebook UUID is required"})
		return
	}

	notebook, err := n.NotebookService.GetNotebookByUUID(notebookUUID)
	if err != nil {
		log.Errorf("Failed to get notebook: %v", err)
		c.JSON(404, gin.H{"error": "Notebook not found"})
		return
	}

	c.JSON(200, notebook)
}

func (n *NotebookAPI) UpdateNotebookByID(c *gin.Context) {
	notebookID := c.Param("id")
	notebook_id, err := strconv.Atoi(notebookID)
	if err != nil || notebook_id == 0 {
		c.JSON(400, gin.H{"error": "Notebook ID is required"})
		return
	}

	var req UpdateNotebookRequest
	if err := c.BindJSON(&req); err != nil {
		log.Infof("Error parsing request body in UpdateNotebookByID. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	// Validate that endtime should be greater than start_time
	if req.StartTime != "" && req.EndTime != "" {
		startTime, err1 := time.Parse(time.RFC3339, req.StartTime)
		endTime, err2 := time.Parse(time.RFC3339, req.EndTime)

		if err1 != nil {
			c.JSON(400, gin.H{"error": "Invalid start_time format. Use RFC3339 format (e.g., 2023-12-25T10:00:00Z)"})
			return
		}

		if err2 != nil {
			c.JSON(400, gin.H{"error": "Invalid end_time format. Use RFC3339 format (e.g., 2023-12-25T15:00:00Z)"})
			return
		}

		if !endTime.After(startTime) {
			c.JSON(400, gin.H{"error": "end_time must be greater than start_time"})
			return
		}
	}

	err = n.NotebookService.UpdateNotebook(notebook_id, req.StartTime, req.EndTime)
	if err != nil {
		log.Errorf("Failed to update notebook: %v", err)
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "Notebook successfully updated."})
}

func (n *NotebookAPI) DeleteNotebookByID(c *gin.Context) {
	notebookID := c.Param("id")
	notebook_id, err := strconv.Atoi(notebookID)
	if err != nil || notebook_id == 0 {
		c.JSON(400, gin.H{"error": "Notebook ID is required"})
		return
	}

	err = n.NotebookService.DeleteNotebook(notebook_id)
	if err != nil {
		log.Errorf("Failed to delete notebook: %v", err)
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(204, gin.H{"msg": "Notebook is deleted."})
}

func (n *NotebookAPI) GetAvailableNotebooks(c *gin.Context) {
	user := c.Param("user_id")
	// string to int
	studentID, err := strconv.Atoi(user)
	if err != nil || studentID == 0 {
		c.JSON(400, gin.H{"error": "student ID is required"})
		return
	}

	mode := c.Query("type") // Get "query" parameter
	var notebookMode int
	var ok bool

	if notebookMode, ok = NotebookMode[mode]; !ok {
		c.JSON(400, gin.H{"error": "invalid notebook mode"})
		return
	}

	notebooks, err := n.NotebookService.GetAvailableNotebooks(notebookMode)
	if err != nil {
		log.Errorf("Failed to get available notebooks: %v", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve available notebooks"})
		return
	}

	// Add to submission status table as status: downloaded
	status := NotebookStatus["downloaded"]
	for _, notebook := range notebooks {
		download_time := time.Now()
		err = n.NotebookService.SaveStudentSubmission(notebook.ID, notebook.Title, notebook.Path, status, &download_time, studentID)
		if err != nil {
			log.Errorf("Failed to update student notebook status: %v", err)
			c.JSON(500, gin.H{"error": "Failed to update student notebook status"})
			return
		}
	}

	c.JSON(200, gin.H{"data": notebooks})
}

func (n *NotebookAPI) ServeNotebookFile(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(400, gin.H{"error": "file_path parameter is required"})
		return
	}

	// Security: Validate file path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	// Validate file extension
	if filepath.Ext(cleanPath) != ".ipynb" {
		c.JSON(400, gin.H{"error": "Only .ipynb files are allowed"})
		return
	}

	// Open file
	file, err := os.Open(cleanPath)
	if err != nil {
		log.Errorf("Failed to open file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// Get file info for proper headers
	fileInfo, err := file.Stat()
	if err != nil {
		log.Errorf("Failed to get file info: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get file info"})
		return
	}

	// Set headers for file download
	fileName := filepath.Base(cleanPath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Copy file content to response
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		log.Errorf("Failed to copy file content: %v", err)
		c.JSON(500, gin.H{"error": "Failed to serve file"})
		return
	}
}

type StudentNotebookSubmitRequest struct {
	Title      string `form:"title" binding:"required"`
	Path       string `form:"path" binding:"required"`
	NotebookID int    `form:"notebook_id" binding:"required"`
	Status     string `form:"status" binding:"required"`
	CreatedAt  string `form:"created_at" binding:"required"`
}

func (n *NotebookAPI) SubmitNotebook(c *gin.Context) {
	var req StudentNotebookSubmitRequest

	user := c.Param("user_id")
	// string to int
	user_id, err := strconv.Atoi(user)
	if err != nil || user_id == 0 {
		c.JSON(400, gin.H{"error": "student ID is required"})
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Errorf("Error parsing JSON data in SubmitNotebook. Err: %v", err)
		c.JSON(400, gin.H{"error": "Invalid JSON data", "details": err.Error()})
		return
	}

	if req.NotebookID == 0 {
		log.Info("Notebook ID cannot be 0.")
		c.JSON(400, gin.H{"error": "Invalid JSON data", "details": "Notebook id cannot be 0."})
		return
	}

	// Check if notebook deadline is missed.
	notebook, err := n.NotebookService.GetNotebookEndTime(req.NotebookID)
	if err != nil {
		log.Infof("Error getting Notebook details. Err: %v", err)
		c.JSON(500, gin.H{"error": "No Notebook found.", "details": err.Error()})
		return
	}

	currentTime := time.Now()
	// Add buffer to the current time
	allowedTime := currentTime.Add(time.Minute * buffer)

	if allowedTime.After(*notebook.EndTime) {
		log.Infof("Error cannot submit notebook after deadline. Student: %v, Notebook: %v ", user_id, notebook.ID)
		c.JSON(500, gin.H{"error": "Cannot submit notebook after deadline."})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("filecontent")
	if err != nil {
		log.Infof("Error getting file from form in UploadNotebook. Err: %v", err)
		c.JSON(400, gin.H{"error": "No file uploaded", "details": err.Error()})
		return
	}

	// Validate file extension
	if filepath.Ext(file.Filename) != ".ipynb" {
		c.JSON(400, gin.H{"error": "File must be a .ipynb notebook file"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := fmt.Sprintf("uploads/submissions/%d/", req.NotebookID)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Errorf("Failed to create uploads directory: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate filename based on title and timestamp
	// timestamp := time.Now().Format("20060102_150405")
	// filename := fmt.Sprintf("%s_%s.ipynb", req.Title, timestamp)

	// Sanitize filename
	filename := filepath.Base(req.Title)
	filePath := filepath.Join(uploadsDir, filename)

	// Save the uploaded file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Errorf("Failed to save uploaded file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// Log the submission
	log.Infof("Notebook submission received - Title: %s, Path: %s, Status: %s Created At: %s",
		req.Title, filePath, req.Status, req.CreatedAt)

	// Parse file creation time
	var fileCreatedAt *time.Time
	if req.CreatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.CreatedAt); err == nil {
			fileCreatedAt = &parsed
		}
	}

	submissionStatus := NotebookStatus[req.Status]

	// Save to database
	err = n.NotebookService.SaveStudentSubmission(req.NotebookID, req.Title, filePath, submissionStatus, fileCreatedAt, user_id)
	if err != nil {
		log.Errorf("Failed to save student submission to database: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save submission to database", "details": err.Error()})
		return
	}

	formattedTime := currentTime.Format("01/02/2006 03:04 PM")

	response := gin.H{
		"message":    fmt.Sprintf("Notebook received at %s.", formattedTime),
		"title":      req.Title,
		"path":       filePath,
		"status":     req.Status,
		"created_at": req.CreatedAt,
	}

	c.JSON(200, response)
}

func (n *NotebookAPI) DownloadNotebooksByID(c *gin.Context) {
	notebookID := c.Param("id")
	notebook_id, err := strconv.Atoi(notebookID)
	if err != nil || notebook_id == 0 {
		c.JSON(400, gin.H{"error": "Notebook ID is required"})
		return
	}

	// Get all submitted notebooks for this assignment
	submissions, err := n.NotebookService.GetSubmittedNotebooksByID(notebook_id)
	if err != nil {
		log.Errorf("Failed to get submitted notebooks: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get submitted notebooks"})
		return
	}

	if len(submissions) == 0 {
		c.JSON(404, gin.H{"error": "No submissions found for this notebook"})
		return
	}

	// Create temporary directory for copying files
	tmpDir := fmt.Sprintf("/tmp/subs_%d_%d", notebook_id, time.Now().Unix())
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		log.Errorf("Failed to create temporary directory: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create temporary directory"})
		return
	}
	defer os.RemoveAll(tmpDir) // Clean up after we're done

	// Copy files to temporary directory
	for _, submission := range submissions {
		if _, err := os.Stat(submission.Path); os.IsNotExist(err) {
			log.Warnf("Submission file not found: %s", submission.Path)
			continue
		}

		// Create filename with user ID prefix
		filename := fmt.Sprintf("%d_%s", submission.UserID, filepath.Base(submission.Path))
		destPath := filepath.Join(tmpDir, filename)

		// Copy file
		err = copyFile(submission.Path, destPath)
		if err != nil {
			log.Errorf("Failed to copy file %s: %v", submission.Path, err)
			continue
		}
	}

	// Create ZIP file
	zipFilename := fmt.Sprintf("submissions_notebook_%d.zip", notebook_id)
	zipPath := filepath.Join(tmpDir, zipFilename)

	err = createZipFile(tmpDir, zipPath)
	if err != nil {
		log.Errorf("Failed to create ZIP file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create ZIP file"})
		return
	}

	// Send ZIP file as response
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
	c.Header("Content-Type", "application/zip")

	c.File(zipPath)
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Helper function to create ZIP file
func createZipFile(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through all files in the source directory
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the ZIP file itself and directories
		if info.IsDir() || filepath.Base(filePath) == filepath.Base(zipPath) {
			return nil
		}

		// Create a ZIP entry
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Copy file content to ZIP entry
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipEntry, file)
		return err
	})
}
