package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func hasFeedbackOnCode(codeFromT, codeFromS string) bool {
	if strings.EqualFold(strings.Replace(codeFromT, " ", "", -1), strings.Replace(codeFromS, " ", "", -1)) {
		return false
	}
	return true
}

// add the middleware function
func appMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")
		fmt.Printf("Auth HEADER: %v\n", authHeader)

		if len(authHeader) != 2 {
			resp := "Malformed Token"
			log.Infof(resp)
			c.JSON(http.StatusUnauthorized, resp)
			return
		}

		token := authHeader[1]
		id := 0
		name := ""
		rows, err := db.Query("select id, name from users where user_uuid = $1", token)
		defer rows.Close()

		if err != nil {
			resp := "Unauthorized. Err: %v"
			log.Infof(resp, err)
			c.JSON(http.StatusUnauthorized, resp)
			return
		}

		for rows.Next() {
			rows.Scan(&id, &name)
		}

		if id == 0 && name == "" {
			resp := "Unauthorized. Err: %v"
			log.Infof(resp, err)
			c.JSON(http.StatusUnauthorized, resp)
			return
		}
		c.Set("user_id", id)

		c.Next()
	}
}
