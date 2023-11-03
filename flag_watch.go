package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type FlagWatchAPI struct {
	FlagWatchService FlagWatchStore
}

func (fwAPI *FlagWatchAPI) FlagSubHandler(c *gin.Context) {
	var sub FlagSubmission

	if err := c.BindJSON(&sub); err != nil {
		log.Infof("Error parsing request body in FlagSubHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	uID, _ := c.Get("user_id")

	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	// sub.Mode = 2 // Mode 2 is for Flag Submissions
	if userID, ok := uID.(int); ok {
		sub.TeacherID = userID
	}

	if sub.ProblemID == 0 || sub.StudentID == 0 || sub.TeacherID == 0 {
		log.Infof("Invalid req body for flag submission. %v\n", sub)
		c.JSON(400, "Failed to flag submission.")
		return
	}

	flagID, err := fwAPI.FlagWatchService.GetFlagWatchSubID(sub.SubmissionID, sub.Mode)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to get existing flag sub for subID. %v Err. %v\n", sub, err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	if flagID != 0 {
		// Update
		sub.ID = flagID
		err = fwAPI.FlagWatchService.UpdateFlagWatchSubs(sub)
		if err != nil {
			log.Infof("Failed to update flag sub. %v Err. %v\n", sub, err)
			c.JSON(500, gin.H{"msg": err})
			return
		}
	} else {
		// Create
		_, err = fwAPI.FlagWatchService.SaveFlagWatchSubs(sub)
		if err != nil {
			log.Infof("Failed to save flag sub. %v Err. %v\n", sub, err)
			c.JSON(500, gin.H{"msg": err})
			return
		}

	}
	c.JSON(http.StatusCreated, gin.H{"msg": "Submission flagged successfully."})

}

func (fwAPI *FlagWatchAPI) GetFlagSubsHandler(c *gin.Context) {
	// uID, _ := c.Get("user_id")

	fSubs, err := fwAPI.FlagWatchService.GetFlagSubs(2, 2)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Flag Submissions in GetFlagSubsHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	c.JSON(200, gin.H{"data": fSubs})

}

func (fwAPI *FlagWatchAPI) DelFlagSubHandler(c *gin.Context) {
	var fSub FlagSubmission
	if err := c.BindJSON(&fSub); err != nil {
		log.Infof("Error parsing request body in DelFlagSubHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err := fwAPI.FlagWatchService.UnFlagWatchSubs(fSub)
	if err != nil {
		log.Infof("Failed to Unflag Submission in DelFlagSubHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	c.JSON(200, gin.H{"msg": "Submission Unflagged successfully."})
}

// Rem
func (fwAPI *FlagWatchAPI) GetWatchSubsHandler(c *gin.Context) {
	// uID, _ := c.Get("user_id")

	fSubs, err := fwAPI.FlagWatchService.GetFlagSubs(1, 1)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Flag Submissions in GetFlagSubsHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	// When there is new snapshot update from the students set on Watch,
	// Update the code block on Watch
	for idx, sub := range fSubs {
		key := fmt.Sprintf("%v-%v", sub.StudentID, sub.ProblemID)
		if val, ok := studentWorkSnapshot[key]; ok {
			fSubs[idx].Code = val.Code
			fSubs[idx].CreatedAt = val.CreatedAt
			fSubs[idx].CreatedAt = val.CreatedAt
		}
	}
	c.JSON(200, gin.H{"data": fSubs})

}
