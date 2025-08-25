package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type FeedbackAgentAPI struct {
	DB         *Database
	HTTPClient *http.Client
	BaseURL    string
}

type AgentFeedback struct {
	SubmissionID int           `json:"submission_id"`
	Model        string        `json:"model"`
	Feedbacks    []interface{} `json:"feedbacks"`
}

func (fa *FeedbackAgentAPI) GetAgentFeedbackByIDHandler(c *gin.Context) {
	id := c.Param("id")
	submissionID, err := strconv.Atoi(id)
	if err != nil || submissionID == 0 {
		log.Infof("Error parsing submission ID in GetAgentFeedbackByIDHandler. Err: %v", err)
		c.JSON(400, gin.H{"msg": "Invalid submission ID"})
		return
	}

	feedback_resp := make([]AgentFeedback, 0)

	// Get submission by ID
	submission, err := fa.DB.GetSubmissionByID(submissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Infof("Submission not found with ID: %v in GetAgentFeedbackByIDHandler", submissionID)
			c.JSON(404, gin.H{"msg": "Submission not found"})
			return
		}
		log.Infof("Failed to get submission by ID in GetAgentFeedbackByIDHandler. Err: %v", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	// Get problem based on submission's problem_id
	problem, err := fa.DB.GetProblemByID(submission.ProblemID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Infof("Problem not found with ID: %v", submission.ProblemID)
			c.JSON(404, gin.H{"msg": "Problem not found"})
			return
		}
		log.Infof("Failed to get problem by ID in GetAgentFeedbackByIDHandler. Err: %v", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	// Get active agents
	all_agents, err := fa.DB.GetAgents()
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to get agents in GetAgentFeedbackByIDHandler. Err: %v", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	for _, agent := range all_agents {
		if agent.IsActive == 1 {
			resp, err := fa.CallExternalFeedbackService(submission, problem, agent)
			if err != nil {
				log.Errorf("Failed to call external feedback service for agent %d: %v", agent.ID, err)
				// Add error response to feedback
				response := AgentFeedback{
					SubmissionID: submission.ID,
					Model:        agent.Model,
					Feedbacks:    []interface{}{"Failed to fetch AI feedback"},
				}
				feedback_resp = append(feedback_resp, response)
				continue
			}

			// Only process successful responses
			if resp != nil {
				response := AgentFeedback{
					SubmissionID: submission.ID,
					Model:        agent.Model,
					Feedbacks:    resp.Feedbacks,
				}
				feedback_resp = append(feedback_resp, response)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": feedback_resp})
}

type ProcessRequest struct {
	Model  string                 `json:"model"`
	Prompt string                 `json:"prompt"`
	Data   map[string]interface{} `json:"data"`
}

type ProcessResponse struct {
	Analysis  string        `json:"analysis"`
	Feedbacks []interface{} `json:"feedbacks"`
}

func (fa *FeedbackAgentAPI) CallExternalFeedbackService(submission Submission, problem Problem, agent Agent) (*ProcessResponse, error) {
	if fa.BaseURL == "" {
		return nil, fmt.Errorf("FEEDBACK_AGENT URL not configured")
	}

	// Prepare data map with submission, problem, and agents
	data := map[string]interface{}{
		"INITIAL_SCAFFOLD": problem.Question,
		"CODE_SNAPSHOT":    submission.Code,
		"TIME_ON_TASK":     time.Now().Sub(problem.CreatedAt).Minutes(),
	}

	// Prepare request payload
	payload := ProcessRequest{
		Model:  agent.Model,
		Prompt: agent.Prompt,
		Data:   data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make POST request to external service
	url := fmt.Sprintf("http://%s/process", fa.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := fa.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external service returned status %d", resp.StatusCode)
	}

	var processResp ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&processResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &processResp, nil
}

type FeedbackRating struct {
	GradeID int `json:"id"`
	Rating  int `json:"rating"`
}

func (fa *FeedbackAgentAPI) UpdateFeedbackRatingHandler(c *gin.Context) {

	var rating FeedbackRating
	if err := c.BindJSON(&rating); err != nil {
		log.Infof("Error parsing request body in UpdateFeedbackRatingHandler. Err: %v", err)
		c.JSON(400, gin.H{"msg": "Invalid request body"})
		return
	}

	err := fa.UpdateFeedbackRating(rating)
	if err != nil {
		log.Infof("Failed to update feedback rating for submission %d. Err: %v", rating.GradeID, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Feedback rating updated successfully",
		"id":  rating.GradeID,
	})
}

func (fa *FeedbackAgentAPI) UpdateFeedbackRating(rating FeedbackRating) error {
	sqlStatement := `UPDATE grades SET rating=$1, updated_at = $2 WHERE id = $3`

	_, err := fa.DB.DB.Exec(sqlStatement,
		rating.Rating,
		time.Now(),
		rating.GradeID)

	if err != nil {
		return fmt.Errorf("failed to update feedback rating: %v", err)
	}

	return nil
}
