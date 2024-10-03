package models

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var AnonymousUser = &User{}

type UserRole = int

const (
	AdminUser UserRole = iota
	FacultyUser
	AccountantUser
	Provc
)

type User struct {
	Id             int
	Username       string
	Email          string
	HashedPassword []byte
	CreatedAt      time.Time
	Role           UserRole
	FullName       string
	Designation    string
	MobileNumber   string
	School         string
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

	// err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	// if err != nil {
	// 	return 0, ErrInvalidCredentials
	// }
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

	err = m.Db.QueryRow("select full_name, designation, mobile_number, school from profile where username = ?", u.Username).Scan(&u.FullName, &u.Designation, &u.MobileNumber, &u.School)
	if err == sql.ErrNoRows {
		return u, nil
	} else if err != nil {
		return nil, err
	}
	return u, nil
}

func (m *UserModel) CreateUser(username, email, password string, role UserRole) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	result, err := m.Db.Exec("INSERT INTO user (username, email, hashed_password, created_at, user_role) VALUES (?, ?, ?, ?, ?)",
		username, email, hashedPassword, time.Now(), role)
	if err != nil {
		return 0, err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(userId), nil
}

func (m *UserModel) ResetPassword(userId int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("UPDATE user SET hashed_password = ? WHERE user_id = ?", hashedPassword, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) GetAdmins() ([]string, error) {

	var admins []string

	rows, err := m.Db.Query("select user_id from user where user_role = 0")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var admin int
		err := rows.Scan(&admin)
		if err != nil {
			return nil, err
		}
		admins = append(admins, strconv.Itoa(admin))
	}

	return admins, nil
}

func (m *UserModel) GetAccounts() ([]string, error) {

	var accounts []string

	rows, err := m.Db.Query("select user_id from user where user_role = 2")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var account int
		err := rows.Scan(&account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, strconv.Itoa(account))
	}

	return accounts, nil
}
