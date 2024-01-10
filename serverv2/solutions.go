package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SolutionAPI struct {
	SolService SolStore
}

func (s *SolutionAPI) SolutionHandler(c *gin.Context) {
	var newSol Solutions

	if err := c.BindJSON(&newSol); err != nil {
		log.Infof("Error parsing request body in SolutionHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	// If the problem is unpublished, the solution should should broadcast.
	expiredProblem, _ := s.SolService.IsExpired(newSol.ProblemID)
	if expiredProblem {
		newSol.Broadcast = 1
		log.Printf("Setting solution to broadcast for problem id: %v", newSol.ProblemID)
	}

	existingSol, err := s.SolService.GetSolution(newSol.ProblemID)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to get existing solution. %v Err. %v\n", newSol, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	if existingSol.ID != 0 {
		newSol.UpdatedAt = time.Now()
		err = s.SolService.UpdateSolution(newSol)
		if err != nil {
			log.Infof("Failed to update solution. %v Err. %v\n", newSol, err)
			c.JSON(500, gin.H{"msg": err.Error()})
			return
		}

	} else {
		newSol.CreatedAt = time.Now()
		newSol.UpdatedAt = time.Now()
		_, err = s.SolService.SaveSolution(newSol)
		if err != nil {
			log.Infof("Failed to save solution. %v Err. %v\n", newSol, err)
			c.JSON(500, gin.H{"msg": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"msg": "Solution successfully uploaded."})

}
