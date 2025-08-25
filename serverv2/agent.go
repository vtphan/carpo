package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AgentAPI struct {
	AgentService AgentStore
}

func (agent *AgentAPI) GetAgentHandler(c *gin.Context) {
	agents, err := agent.AgentService.GetAgents()
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Get Agents in GetAgentHandler. Err. %v\n", err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

func (agent *AgentAPI) SaveAgentHandler(c *gin.Context) {
	var newAgent Agent

	if err := c.BindJSON(&newAgent); err != nil {
		log.Infof("Error parsing request body in SaveAgentHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	newAgent.CreatedAt = time.Now()
	newAgent.UpdatedAt = time.Now()
	newAgent.IsActive = 1

	id, err := agent.AgentService.CreateAgent(newAgent)
	if err != nil && err != sql.ErrNoRows {
		log.Infof("Failed to Save Agent. %v Err. %v\n", id, err)
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	newAgent.ID = id
	c.JSON(http.StatusCreated, newAgent)
}

func (agent *AgentAPI) UpdateAgentHandler(c *gin.Context) {
	id := c.Param("id")
	agentID, err := strconv.Atoi(id)
	if err != nil || agentID == 0 {
		log.Infof("Error parsing request param in UpdateAgentHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	var updatedAgent Agent
	if err := c.BindJSON(&updatedAgent); err != nil {
		log.Infof("Error parsing request body in UpdateAgentHandler. Err: %v", err)
		c.JSON(400, err.Error())
		return
	}

	updatedAgent.ID = agentID
	updatedAgent.UpdatedAt = time.Now()

	err = agent.AgentService.UpdateAgent(updatedAgent)
	if err != nil {
		log.Printf("Failed to update agent with ID: %v. Err: %v", agentID, err)
		c.JSON(500, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": agentID, "msg": "Updated Agent successfully."})
}

func (agent *AgentAPI) DeleteAgentHandler(c *gin.Context) {
	id := c.Param("id")
	agentID, err := strconv.Atoi(id)
	if err != nil || agentID == 0 {
		log.Infof("Error parsing request param in DeleteAgentHandler. Err: %v", err)
		c.JSON(400, err)
		return
	}

	err = agent.AgentService.DeleteAgent(agentID)
	if err != nil {
		log.Printf("Failed to delete agent with ID: %v. Err: %v", agentID, err)
		c.JSON(500, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": agentID, "msg": "Agent deleted successfully."})
}