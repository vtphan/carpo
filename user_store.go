package main

import (
	"database/sql"
)

type UserStore interface {
	SaveUser(string, string, int) (int, error)
	GetUserByName(string) (User, error)
	UpdateUUID(User) error
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Role int    `json:"role"`
}

type Database struct {
	DB *sql.DB
}

func (db *Database) SaveUser(name string, uuid string, role int) (id int, err error) {

	sqlStatement := `INSERT INTO users (name, user_uuid, role) values ($1, $2, $3) RETURNING id;`

	err = db.DB.QueryRow(sqlStatement, name, uuid, role).Scan(&id)

	if err != nil {
		return id, err
	}
	return
}

func (db *Database) GetUserByName(name string) (u User, err error) {

	sqlStatement := `SELECT id, name, user_uuid, role FROM users WHERE name=$1 LIMIT 1;`

	row := db.DB.QueryRow(sqlStatement, name)
	err = row.Scan(&u.ID, &u.Name, &u.UUID, &u.Role)

	if err != nil {
		return u, err
	}
	return

}

func (db *Database) GetUserNameByID(id int) (u User, err error) {

	sqlStatement := `SELECT id, name, user_uuid, role FROM users WHERE id=$1 LIMIT 1;`

	row := db.DB.QueryRow(sqlStatement, id)
	err = row.Scan(&u.ID, &u.Name, &u.UUID, &u.Role)

	if err != nil {
		return u, err
	}
	return

}

func (db *Database) UpdateUUID(user User) (err error) {
	sqlStatement := `UPDATE users SET user_uuid = $2 WHERE id = $1;`
	_, err = db.DB.Exec(sqlStatement, user.ID, user.UUID)
	if err != nil {
		return err
	}
	return
}
