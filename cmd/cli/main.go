package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sporic/sporic/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN not provided in .env")
	}

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping the database: %v", err)
	}
}

func main() {
	createUserFlag := flag.NewFlagSet("createUser", flag.ExitOnError)
	username := createUserFlag.String("username", "", "Username")
	email := createUserFlag.String("email", "", "Email")
	password := createUserFlag.String("password", "", "Password")
	role := createUserFlag.Int("role", 0, "Role (0 = Admin, 1 = Faculty)")

	resetPasswordFlag := flag.NewFlagSet("resetPassword", flag.ExitOnError)
	userId := resetPasswordFlag.Int("user-id", 0, "User ID")
	newPassword := resetPasswordFlag.String("password", "", "New Password")

	if len(os.Args) < 2 {
		fmt.Println("create_user or reset_password subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create_user":
		createUserFlag.Parse(os.Args[2:])
		if *username == "" || *email == "" || *password == "" {
			fmt.Println("All fields (username, email, password, and role) are required.")
			os.Exit(1)
		}
		err := createUser(*username, *email, *password, models.UserRole(*role))
		if err != nil {
			log.Fatalf("Error creating user: %v", err)
		}
	case "reset_password":
		resetPasswordFlag.Parse(os.Args[2:])
		if *userId == 0 || *newPassword == "" {
			fmt.Println("Both user-id and password are required.")
			os.Exit(1)
		}
		err := resetPassword(*userId, *newPassword)
		if err != nil {
			log.Fatalf("Error resetting password: %v", err)
		}
	default:
		fmt.Println("Unknown command. Use 'createUser' or 'resetPassword'.")
		os.Exit(1)
	}
}

func createUser(username, email, password string, role models.UserRole) error {
	userModel := &models.UserModel{Db: db}

	userId, err := userModel.CreateUser(username, email, password, role)
	if err != nil {
		return err
	}
	fmt.Printf("User created successfully with ID: %d\n", userId)
	return nil
}

func resetPassword(userId int, newPassword string) error {
	userModel := &models.UserModel{Db: db}

	err := userModel.ResetPassword(userId, newPassword)
	if err != nil {
		return err
	}
	fmt.Println("Password reset successfully")
	return nil
}
