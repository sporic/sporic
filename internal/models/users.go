package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id             int
	Username       string
	Email          string
	HashedPassword []byte
	CreatedAt      time.Time
}

type UserModel struct {
	Db *sql.DB
}

func (m *UserModel) Authenticate(username string, password string) (int, error) {

	var id int
	var hashedPassword []byte
	err := m.Db.QueryRow("select id, hashed_password from users where username = ?", username).Scan(&id, &hashedPassword)
	if err==sql.ErrNoRows{
		
	}else
	{

	}
	return id, err
}
