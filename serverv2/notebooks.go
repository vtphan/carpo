package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type NotebookAPI struct {
	NotebookService NotebookStore
}

type UploadNotebookRequest struct {
	Title         string `form:"title" binding:"required"`
	Mode          int    `form:"mode" binding:"required"`
	AvailableTill string `form:"available_till"`
	EndTime       string `form:"end_time"`
	UserID        int    `form:"user_id" binding:"required"`
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

	// Parse timestamp strings
	var availableTill, endTime *time.Time
	if req.AvailableTill != "" {
		if parsed, err := time.Parse(time.RFC3339, req.AvailableTill); err == nil {
			availableTill = &parsed
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
	filename := fmt.Sprintf("%s_%s.ipynb", req.Title, timestamp)
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
		availableTill,
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
		"notebook_uuid":  notebookUUID,
		"title":          req.Title,
		"mode":           req.Mode,
		"path":           filePath,
		"available_till": availableTill,
		"end_time":       endTime,
		"user_id":        req.UserID,
		"message":        "Notebook uploaded successfully",
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

func (n *NotebookAPI) GetAvailableNotebooks(c *gin.Context) {
	studentID := c.Param("user_id")
	if studentID == "" {
		c.JSON(400, gin.H{"error": "student ID is required"})
		return
	}

	notebooks, err := n.NotebookService.GetAvailableNotebooks()
	if err != nil {
		log.Errorf("Failed to get available notebooks: %v", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve available notebooks"})
		return
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
	uploadsDir := "uploads/students/notebooks"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Errorf("Failed to create uploads directory: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate filename based on title and timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.ipynb", req.Title, timestamp)
	// Sanitize filename
	filename = filepath.Base(filename)
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

	// For now, use notebook_id as 1 (we need to find the actual notebook by title later)
	// Submission status: 1 = submitted
	submissionStatus := 1

	// Save to database
	err = n.NotebookService.SaveStudentSubmission(req.NotebookID, req.Title, filePath, submissionStatus, fileCreatedAt, user_id)
	if err != nil {
		log.Errorf("Failed to save student submission to database: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save submission to database", "details": err.Error()})
		return
	}

	response := gin.H{
		"message":    "Notebook submission received and saved successfully.",
		"title":      req.Title,
		"path":       filePath,
		"status":     req.Status,
		"created_at": req.CreatedAt,
	}

	c.JSON(200, response)
}
