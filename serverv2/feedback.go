package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type FeedbackAPI struct {
	FeedbackService FeedbackStore
}

func (fback *FeedbackAPI) GetFeedbackHandler(c *gin.Context) {
	user := c.Param("user_id")
	// string to int
	user_id, err := strconv.Atoi(user)
	if err != nil || user_id == 0 {
		log.Infof("Error parsing user_id params in GetFeedbackHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	problem := c.Param("problem_id")
	// string to int
	problem_id, err := strconv.Atoi(problem)
	if err != nil || user_id == 0 {
		log.Infof("Error parsing problem_id params in GetFeedbackHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	feedbacks, err := fback.FeedbackService.GetStudentFeedbackFromPID(user_id, problem_id)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Feedbacks in GetFeedbackHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": feedbacks})

}
