package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/sporic/sporic/internal/mailer"
)

type NotificationModel struct {
	Db *sql.DB
}

type NotificationType = int

const (
	NewProjectApproval        NotificationType = iota //
	NewExpenditureApproval                            //
	NewInvoiceRequestApproval                         //
	CompletionProjectApproval                         //
	ProjectDelayed                                    // to admins and faculty
	PaymentDelayed                                    // to admins and faculty
	ApplicationApproved                               //
	ApplicationRejected                               //
	ApplicationCompleted
	ExpenditureApprovedNotification //
	ExpenditureRejectedNotification //
	PaymentApprovedNotification     //
	PaymentRejectedNotification
	PaymentInvoiceRequest //
	PaymentApproval       //
	ExpenditurePaid
)

var NotificationTypeMap = map[NotificationType]string{
	NewProjectApproval:              "New Project %s has been submitted for approval",
	NewExpenditureApproval:          "New Expenditure for project %s has been submitted for approval",
	NewInvoiceRequestApproval:       "New Invoice Request for project %s has been submitted for approval",
	CompletionProjectApproval:       "Project %s has been marked for completion approval",
	ProjectDelayed:                  "Project %s has been delayed",
	PaymentDelayed:                  "Payment %s for project %s has been delayed",
	ApplicationApproved:             "Project %s has been approved",
	ApplicationRejected:             "Project %s has been rejected",
	ApplicationCompleted:            "Project %s has been completed",
	ExpenditureApprovedNotification: "Expenditure for project %s has been approved",
	ExpenditureRejectedNotification: "Expenditure for project %s has been rejected",
	PaymentApprovedNotification:     "Payment %s for project %s has been approved by accounts",
	PaymentRejectedNotification:     "Payment %s for project %s has been rejected by accounts",
	PaymentInvoiceRequest:           "Invoice Request for project %s has been submitted",
	PaymentApproval:                 "Payment proof %s for project %s has been uploaded",
	ExpenditurePaid:                 "Expenditure  %s for project %s has been paid",
}

type Notification struct {
	CreatedAt   time.Time
	NotiType    NotificationType
	Description string
	To          []string
}

func (n *NotificationModel) SendNotification(notification Notification, mailer mailer.Mailer) error {
	for _, user := range notification.To {
		_, err := n.Db.Exec("insert into notifications (craeted_at, notification_type, notification_description, notification_to)values (?,?,?,?)", notification.CreatedAt, notification.NotiType, notification.Description, user)

		if err != nil {
			return err
		}
	}

	for _, user := range notification.To {
		row, _ := n.Db.Query("select email from user where user_id = ? ", user)
		var email string
		err := row.Scan(&email)
		if err != nil {
			return err
		}
		err = mailer.Send(email, "example.tmpl", notification)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (n *NotificationModel) RecieveNotification(receivers []string) ([]Notification, error) {
	var notifications []Notification

	for _, username := range receivers {
		rows, err := n.Db.Query("select craeted_at, notification_type, notification_description from notifications where notification_to = ?", username)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var notification Notification
			err = rows.Scan(&notification.CreatedAt, &notification.NotiType, &notification.Description)
			if err != nil {
				return nil, err
			}
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}
