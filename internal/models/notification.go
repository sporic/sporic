package models

import (
	"database/sql"
	"fmt"
	"time"
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
	PaymentApprovedNotification:     "Payment %s for project %s has been approved",
	PaymentRejectedNotification:     "Payment %s for project %s has been rejected",
	PaymentInvoiceRequest:           "Invoice Request for project %s has been submitted",
	PaymentApproval:                 "Payment %s for project %s has been approved by accounts",
	ExpenditurePaid:                 "Expenditure  %s for project %s has been paid",
}

type Notification struct {
	CreatedAt   time.Time
	NotiType    NotificationType
	Description string
	To          []string
}

func (n *NotificationModel) SendNotification(notification Notification) error {
	fmt.Println("4")
	for _, user := range notification.To {
		fmt.Println("5")
		_, err := n.Db.Exec("insert into notifications (craeted_at, notification_type, notification_description, notification_to)values (?,?,?,?)", notification.CreatedAt, notification.NotiType, notification.Description, user)

		if err != nil {
			return err
		}
	}
	fmt.Println("6")
	return nil
}

func (n *NotificationModel) RecieveNotification(username []string) ([]Notification, error) {
	var notifications []Notification

	rows, err := n.Db.Query("select craeted_at, notification_type, notification_description where notification_to = ?", username)
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

	return notifications, nil
}
