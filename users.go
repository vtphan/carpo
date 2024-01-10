package main

import (
	"database/sql"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type UserAPI struct {
	UserService UserStore
	// ProblemService ProblemStore
}

func (u *UserAPI) RegisterUser(c *gin.Context) {
	var newUser User

	if err := c.BindJSON(&newUser); err != nil {
		log.Infof("Error parsing request body in RegisterUser. Err: %v", err)
		c.JSON(400, err)
		return
	}

	if newUser.Name == "" {
		resp := "Name cannot be empty."
		log.Infof(resp)
		c.JSON(422, gin.H{"msg": resp})
		return
	}

	if newUser.Role == 0 {
		resp := "Role cannot be 0."
		log.Infof(resp)
		c.JSON(422, gin.H{"msg": resp})
		return
	}

	user, err := u.UserService.GetUserByName(newUser.Name)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Error getting user by name. Err: %v", err)
		c.JSON(500, gin.H{"msg": err})
		return
	}

	newUser.UUID = uuid.New().String()

	// When newUser.Role == 1, make sure they are in the ENV variables:
	if newUser.Role == 1 && !(slices.Contains(strings.Split(os.Getenv("TA"), ","), newUser.Name) || slices.Contains(strings.Split(os.Getenv("INS"), ","), newUser.Name)) {
		log.Infof("Error registering user as INS or TA.")
		c.JSON(500, gin.H{"msg": "Error registering user as INS or TA."})
		return

	}

	if user.ID != 0 {
		//update already existing user
		user.UUID = newUser.UUID
		err := u.UserService.UpdateUUID(user)
		if err != nil {
			log.Infof("Failed to update user %v. Err: %v", user.Name, err)
			c.JSON(500, gin.H{"msg": err})
			return

		}
		c.JSON(200, user)

	} else {
		// create new user
		id, err := u.UserService.SaveUser(newUser.Name, newUser.UUID, newUser.Role)
		if err != nil {
			log.Infof("Error in saving user. Err: %v", err)
			c.JSON(500, gin.H{"msg": err})
			return
		}
		newUser.ID = id
		c.JSON(200, newUser)
	}

}
