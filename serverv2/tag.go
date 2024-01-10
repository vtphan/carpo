package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type TagAPI struct {
	TagService TagStore
}

func (tag *TagAPI) GetTagHandler(c *gin.Context) {

	m := c.Query("mode")
	// string to int
	mode, err := strconv.Atoi(m)
	if err != nil || mode == 0 {
		log.Infof("Error parsing request query in GetTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	tags, err := tag.TagService.GetTags(mode)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Tags in GetTagHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})

}

func (tag *TagAPI) GetAllTagHandler(c *gin.Context) {

	tags, err := tag.TagService.GetAllTags()
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get all Tags in GetAllTagHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})

}

func (tag *TagAPI) SaveTagHandler(c *gin.Context) {
	var newTag Tag

	if err := c.BindJSON(&newTag); err != nil {
		log.Infof("Error parsing request body in SaveTagHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	uID, _ := c.Get("user_id")

	if userID, ok := uID.(int); ok {
		newTag.UserID = userID
	}
	// else {
	// 	c.AbortWithStatus(http.StatusUnauthorized)
	// 	return
	// }
	newTag.CreatedAt = time.Now()
	newTag.UpdatedAt = time.Now()
	newTag.Status = 1

	fmt.Printf("UserID: %v, Request: %+v", uID, newTag)

	id, err := tag.TagService.CreateTag(newTag)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Save Submission. %v Err. %v\n", id, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	newTag.ID = id
	c.JSON(http.StatusCreated, newTag)

}

func (tag *TagAPI) DeleteTagHandler(c *gin.Context) {
	id := c.Param("id")
	// string to int
	tagID, err := strconv.Atoi(id)
	if err != nil || tagID == 0 {
		log.Infof("Error parsing request body in DeleteTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err = tag.TagService.DeleteTag(tagID)
	if err != nil {
		log.Printf("Failed to delete tag with ID: %v. Err: %v", tagID, err)
		c.JSON(500, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": tagID, "msg": "Tag archived successfully."})
}

func (tag *TagAPI) UpdateTagHandler(c *gin.Context) {
	id := c.Param("id")
	// string to int
	tagID, err := strconv.Atoi(id)
	if err != nil || tagID == 0 {
		log.Infof("Error parsing request body in UpdateTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	var Tag Tag
	if err := c.BindJSON(&Tag); err != nil {
		log.Infof("Error parsing request body in UpdateTagHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	err = tag.TagService.UpdateTagName(Tag.Name, tagID)
	if err != nil {
		log.Printf("Failed to update tag name with ID: %v. Err: %v", tagID, err)
		c.JSON(500, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": tagID, "msg": "Updated Tag successfully."})
}

func (tag *TagAPI) TagProblemHandler(c *gin.Context) {
	var newTagP TagProblem

	if err := c.BindJSON(&newTagP); err != nil {
		log.Infof("Error parsing request body in TagProblemHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	uID, _ := c.Get("user_id")

	if userID, ok := uID.(int); ok {
		newTagP.UserID = userID
	}
	newTagP.CreatedAt = time.Now()

	err := tag.TagService.SaveProblemTag(newTagP)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Save ProblemTag %v. Err. %v\n", newTagP, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newTagP)

}

func (tag *TagAPI) TagProblemDelHandler(c *gin.Context) {
	tID := c.Param("id")
	pID := c.Param("pid")
	// string to int
	tagID, err := strconv.Atoi(tID)
	if err != nil || tagID == 0 {
		log.Infof("Error parsing tag id in DeleteTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	probID, err := strconv.Atoi(pID)
	if err != nil || probID == 0 {
		log.Infof("Error parsing prob id in DeleteTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err = tag.TagService.DeleteProblemTag(tagID, probID)
	if err != nil {
		log.Infof("Failed to Delete ProblemTag %v %v. Err. %v\n", tagID, probID, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Problem Tag removed successfully."})
}

func (tag *TagAPI) TagSubmissionHandler(c *gin.Context) {
	var newTagS TagSubmission

	if err := c.BindJSON(&newTagS); err != nil {
		log.Infof("Error parsing request body in TagSubmissionHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	uID, _ := c.Get("user_id")

	if userID, ok := uID.(int); ok {
		newTagS.UserID = userID
	}
	newTagS.CreatedAt = time.Now()

	err := tag.TagService.SaveSubmissionTag(newTagS)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Save SubmissionTag %v. Err. %v\n", newTagS, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newTagS)
}

func (tag *TagAPI) TagSubmissionDelHandler(c *gin.Context) {
	tID := c.Param("id")
	sID := c.Param("sid")
	// string to int
	tagID, err := strconv.Atoi(tID)
	if err != nil || tagID == 0 {
		log.Infof("Error parsing tag id in DeleteTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	subID, err := strconv.Atoi(sID)
	if err != nil || subID == 0 {
		log.Infof("Error parsing sub id in DeleteTagHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err = tag.TagService.DeleteSubmissionTag(tagID, subID)
	if err != nil {
		log.Infof("Failed to Delete SubmissionTag %v %v. Err. %v\n", tagID, subID, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Submission Tag removed successfully."})
}
