package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var AnonymousUser = &User{}

type UserRole = int

const (
	AdminUser UserRole = iota
	FacultyUser
)

type User struct {
	Id             int
	Username       string
	Email          string
	HashedPassword []byte
	CreatedAt      time.Time
	Role           UserRole
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserModel struct {
	Db *sql.DB
}

func (m *UserModel) Authenticate(username string, password string) (int, error) {

	var id int
	var hashedPassword []byte
	err := m.Db.QueryRow("select user_id, hashed_password from user where username = ?", username).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return 0, ErrInvalidCredentials
	}
	return id, nil
}

func (m *UserModel) Get(id int) (*User, error) {
	u := &User{}
	err := m.Db.QueryRow("select user_id, username, email, hashed_password, created_at, user_role from user where user_id = ?", id).Scan(&u.Id, &u.Username, &u.Email, &u.HashedPassword, &u.CreatedAt, &u.Role)
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}
	return u, nil
}
