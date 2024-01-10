package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SubmissionAPI struct {
	SubService SubStore
}

func (sub *SubmissionAPI) SubmissionHandler(c *gin.Context) {
	var newSub Submission

	user := c.Param("user_id")
	// string to int
	user_id, err := strconv.Atoi(user)
	if err != nil || user_id == 0 {
		log.Infof("Error parsing request params in SubmissionHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	if err := c.BindJSON(&newSub); err != nil {
		log.Infof("Error parsing request body in SubmissionHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	if newSub.ProblemID == 0 || newSub.StudentID == 0 {
		log.Infof("Invalid req body for sub_type %v with p_id %v from student id %v.", newSub.Snapshot, newSub.ProblemID, newSub.StudentID)
		c.JSON(400, "Failed to save submission/snapshot.")
		return
	}
	newSub.CreatedAt = time.Now()
	newSub.UpdatedAt = time.Now()

	studentKey := fmt.Sprintf("%v-%v", newSub.StudentID, newSub.ProblemID)
	// Ignore snapshot update if problem is expired
	expiredProb, err := sub.SubService.IsExpired(newSub.ProblemID)
	if err != nil {
		resp := "Failed to check problem expiry. Err: %v"
		log.Infof(resp, err)
		c.JSON(500, gin.H{"msg": resp})
		return
	}
	// Ignore snapshot if the problem is expired
	if expiredProb && newSub.Snapshot == 1 {
		log.Printf("Discard snapshot for inactive problem with key: %s.", studentKey)
		c.JSON(200, gin.H{"msg": "Snapshot no longer needed."})
		return
	}

	// Update
	if val, ok := studentWorkSnapshot[studentKey]; ok {
		// Check for codesnapshot
		if val.Code == newSub.Code && newSub.Snapshot == 1 {
			log.Printf("No change for Student Key: %s.", studentKey)
			c.JSON(200, gin.H{"msg": "No new change found."})
			return
		}
	}

	// check if the newSub is allowed.
	if newSub.Snapshot == 2 {
		if !newSub.IsAllowed() {
			log.Printf("Submission is not allowed within 30 seconds of previous submission. StudentID: %v\n", newSub.StudentID)
			c.JSON(500, gin.H{"msg": "Please wait for 30 seconds before you make another submission on this problem."})
			return
		}
	}

	s, err := sub.SubService.SaveSubmission(newSub)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Save Submission. %v Err. %v\n", s, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	// Update studentWorkSnapshot from Submission that is saved to DB.
	newSub.ID = s
	if !expiredProb {
		stuWrkSnapshotMutex.Lock()
		defer stuWrkSnapshotMutex.Unlock()
		studentWorkSnapshot[studentKey] = newSub
	}

	// update the last submission time for the student
	// skip the update for snapshot case
	if newSub.Snapshot == 2 {
		studentLastSubmission[user_id] = newSub.CreatedAt
	}

	c.JSON(200, gin.H{"msg": "Submission saved successfully."})

}

func (sub *SubmissionAPI) GetSubmissionsHandler(c *gin.Context) {
	submissions, err := sub.SubService.GetSubmissions()
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Submissions in GetSubmissionsHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": submissions})
}

// TODO: sort by Name and creation time
func (sub *SubmissionAPI) GetSnapshotsHandler(c *gin.Context) {
	snapshots := make([]Submission, 0)

	for _, value := range studentWorkSnapshot {
		s := Submission{}

		user, err := sub.SubService.GetUserNameByID(value.StudentID)
		if err != nil {
			log.Infof("Failed to get Name from ID in GetSnapshotsHandler. Err. %v\n", err)
			c.JSON(500, gin.H{"msg": err.Error()})
			return
		}

		s.ID = value.ID
		s.Code = value.Code
		s.Message = value.Message
		s.StudentID = value.StudentID
		s.Name = user.Name
		s.ProblemID = value.ProblemID
		s.CreatedAt = value.CreatedAt

		snapshots = append(snapshots, s)
	}

	c.JSON(200, gin.H{"data": snapshots})
}
