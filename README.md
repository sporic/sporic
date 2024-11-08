
# Sporic

## Table of Contents

- [Features](#features)
- [Configuration](#configuration)
- [Installation](#installation)
- [Usage](#usage)

## Features

### 1. Admin Dashboard
- Role-based access for admins to manage projects efficiently.
- Status tracking for projects, with approval workflows.
- Functionality to export project data to Excel for analysis and reporting.

### 2. Faculty Portal
- Allows faculty members to update project information, upload relevant documents, and receive real-time notifications.
- Enhances record accuracy and engagement from faculty members.

### 3. Accounts Module
- Tracks project expenditures and payment statuses, organizing financial data and transaction history for each project.
- Provides insights into project finances, supporting better financial management.

### 4. Status Management
- Manages the lifecycle of both projects and payments, ensuring precise oversight at each stage.
- Improves overall efficiency by allowing detailed status updates and tracking across all activities.

## Configuration

Before running the application, set up your environment variables in a `.env` file based on the provided `.env.sample`.

| Variable      | Description                                                                                       | Example           g                      |
|---------------|---------------------------------------------------------------------------------------------------|-----------------------------------------|
| `DSN`         | Connection string for the MySQL database, including username, password, host, and database name.  | `user:pass@tcp(localhost:3306)/db?parseTime=true` |
| `ADDR`        | Address and port for running the application server.                                              | `:8080`                                 |
| `SMTP_HOST`   | SMTP server address for sending email notifications.                                              | `smtp.example.com`                      |
| `SMTP_PORT`   | Port for connecting to the SMTP server.                                                           | `25`                                    |
| `SMTP_USER`   | Username for authenticating with the SMTP server.                                                 | `username`                              |
| `SMTP_PASS`   | Password for authenticating with the SMTP server.                                                 | `password`                              |
| `SMTP_SENDER` | Display name and email address for the sender in email notifications.                             | `Name <no-reply@example.com>`           |

Replace placeholders with actual values when setting up the environment.

## Installation

To run Sporic locally, follow these steps:

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/your-username/sporic.git
   cd sporic
   ```

2. **Set Up the Database:**
   - Create a mysql database.
   - Update the `.env` file with your database credentials.

3. **Run Migrations with golang-migrate:**
   - [Install golang-migrate](https://github.com/golang-migrate/migrate) if you haven’t already.
   - Apply migrations to set up the database schema:
     ```bash
     migrate -path ./migrations -database "$DSN" up
     ```
## Usage

1. Start the application by running:
   ```bash
   go run ./cmd/web
   ```

2. Access the application in your browser.
