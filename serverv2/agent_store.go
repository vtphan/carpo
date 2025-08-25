package main

import (
	"time"
)

type Agent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prompt      string    `json:"prompt"`
	Model       string    `json:"model"`
	IsActive    int       `json:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AgentStore interface {
	GetAgents() ([]Agent, error)
	CreateAgent(Agent) (int, error)
	UpdateAgent(Agent) error
	DeleteAgent(int) error
}

func (db *Database) GetAgents() ([]Agent, error) {
	agents := make([]Agent, 0)
	rows, err := db.DB.Query("SELECT id, name, description, prompt, model, is_active, created_at, updated_at FROM agent_settings ORDER BY created_at DESC")
	if err != nil {
		return agents, err
	}
	defer rows.Close()

	for rows.Next() {
		a := Agent{}
		err = rows.Scan(&a.ID, &a.Name, &a.Description, &a.Prompt, &a.Model, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return agents, err
		}
		agents = append(agents, a)
	}
	return agents, err
}

func (db *Database) CreateAgent(agent Agent) (id int, err error) {
	sqlStatement := `INSERT INTO agent_settings (name, description, prompt, model, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = db.DB.QueryRow(sqlStatement, agent.Name, agent.Description, agent.Prompt, agent.Model, agent.IsActive, agent.CreatedAt, agent.UpdatedAt).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) UpdateAgent(agent Agent) (err error) {
	sqlStatement := `UPDATE agent_settings SET name = $1, description = $2, prompt = $3, model = $4, is_active=$5, updated_at = $6 WHERE id = $7`

	_, err = db.DB.Exec(sqlStatement, agent.Name, agent.Description, agent.Prompt, agent.Model, agent.IsActive, agent.UpdatedAt, agent.ID)

	if err != nil {
		return err
	}
	return
}

func (db *Database) DeleteAgent(agentID int) (err error) {
	sqlStatement := `DELETE from agent_settings WHERE id = $1`

	_, err = db.DB.Exec(sqlStatement, agentID)

	if err != nil {
		return err
	}
	return
}
