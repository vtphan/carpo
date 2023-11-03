package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ProblemAPI struct {
	ProblemService ProblemStore
}

func (p *ProblemAPI) PublishProblem(c *gin.Context) {
	var newProblem Problem

	if err := c.BindJSON(&newProblem); err != nil {
		log.Infof("Error parsing request body in PublishProblem. Err: %v", err)
		c.JSON(400, err)
		return
	}

	// QuestionLife defaults to 90 minutes and status is Active (1)
	newProblem.Lifetime = time.Now().Add((time.Minute * time.Duration(90)))

	// create new user
	id, err := p.ProblemService.SaveProblem(newProblem.UserID, newProblem.Question, newProblem.Format, newProblem.Lifetime)
	if err != nil {
		log.Infof("Error in saving new problem. Err: %v", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}
	newProblem.ID = id
	c.JSON(200, newProblem)
}

func (p *ProblemAPI) GetActiveProblems(c *gin.Context) {
	user := c.Param("user_id")
	// string to int
	user_id, err := strconv.Atoi(user)
	if err != nil || user_id == 0 {
		log.Infof("Error parsing request body in GetActiveProblems. Err: %v", err)
		c.JSON(400, err)
		return
	}

	activeProblems, err := p.ProblemService.GetProblems()
	if err != nil {
		log.Infof("Error in saving new problem. Err: %v", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activeProblems})

}

func (p *ProblemAPI) UnpublishProblem(c *gin.Context) {
	problem := c.Param("id")
	// string to int
	problemID, err := strconv.Atoi(problem)
	if err != nil || problemID == 0 {
		log.Infof("Error parsing request body in UnpublishProblem. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err = p.ProblemService.ArchiveProblem(problemID)
	if err != nil {
		log.Printf("Failed to archive Problem ID: %v. Err: %v", problemID, err)
		c.JSON(500, err)
		return
	}

	// Remove students work from studentWorkSnapshot
	for k := range studentWorkSnapshot {
		expiredProblem := fmt.Sprintf("-%d", problemID)
		if strings.Contains(k, expiredProblem) {
			log.Printf("Deleting student Work Snapshot from map with key: %s.", k)
			delete(studentWorkSnapshot, k)
		}
	}

	c.JSON(http.StatusOK, gin.H{"id": problemID, "msg": "Question archived successfully."})
}
