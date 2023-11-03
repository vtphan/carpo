package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type GradeAPI struct {
	GradeService GradeStore
}

func (grade *GradeAPI) GradeHandler(c *gin.Context) {
	var newGrade GradeFeedback

	if err := c.BindJSON(&newGrade); err != nil {
		log.Infof("Error parsing request body in GradeHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	uID, _ := c.Get("user_id")

	newGrade.CreatedAt = time.Now()
	newGrade.UpdatedAt = time.Now()
	if userID, ok := uID.(int); ok {
		newGrade.TeacherID = userID
	}

	// Check if the code is different from student's submission
	studentCode, err := grade.GradeService.GetSubCodeFromID(newGrade.SubmissionID)
	if err != nil {
		resp := "Failed to get student code from sub id. Err: %v"
		log.Infof(resp, err)
		c.JSON(500, gin.H{"msg": resp})
		return
	}

	if hasFeedbackOnCode(newGrade.Code, studentCode) {
		newGrade.HasFeedback = 1
		newGrade.FeedbackAt = time.Now()
	}

	existingGrade, err := grade.GradeService.GetSubGrade(newGrade.SubmissionID)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to get existing grade. %v Err. %v\n", newGrade, err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	if existingGrade.ID != 0 {
		err = grade.GradeService.UpdateGradeFeedback(newGrade)
		if err != nil {
			log.Infof("Failed to update grade. %v Err. %v\n", newGrade, err)
			c.JSON(500, gin.H{"msg": err})
			return
		}

	} else {

		_, err = grade.GradeService.SaveGradeFeedback(newGrade)
		if err != nil {
			log.Infof("Failed to save grade. %v Err. %v\n", newGrade, err)
			c.JSON(500, gin.H{"msg": err})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"msg": "Submission graded successfully."})
}
