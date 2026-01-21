package notification

import (
	"database/sql"
	"encoding/json"
)

// Create new notification - usually run in conjunction with flag setting so runs as a transaction
func (r *Repository) queueNotification(tx *sql.Tx, n NewNotification) error {
	query := "INSERT INTO notification_queue (user_id, message, title, time_added) VALUES ($1, $2, $3, $4)"

	if _, err := tx.Exec(query, n.UserID, n.Msg, n.Title, n.Time); err != nil {
		return err
	}

	return nil
}

// Find all pending notifications in queue and match to user config
func (r *Repository) getPendingNotifications() ([]Notification, error) {
	query := `SELECT nq.user_id, nq.id, nq.message, nq.title, un.type, un.config
			FROM notification_queue AS nq
			INNER JOIN user_notifications AS un
    		ON nq.user_id = un.user_id;`

	rows, err := r.Base.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notifications := []Notification{}

	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.UserID, &n.Id, &n.Msg, &n.Title, &n.Type, &n.Cfg); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

// Get current user notification settings
func (r *Repository) getUserNotificationSettings(userID int64) (NotificationConfigResponse, error) {
	var n NotificationConfigResponse
	var cfgStr string

	if err := r.Base.DB.QueryRow("SELECT type, config FROM user_notifications WHERE user_id = $1", userID).Scan(&n.Type, &cfgStr); err != nil {
		return NotificationConfigResponse{}, err
	}

	// Unmarshal and populate struct
	if err := json.Unmarshal([]byte(cfgStr), &n.ChannelConfig); err != nil {
		return NotificationConfigResponse{}, err
	}

	return n, nil
}

// Add or update user notification settings
func (r *Repository) setUserNotificationSettings(userID int64, notifType string, channelConfig string) error {
	query := `INSERT INTO user_notifications (user_id, type, config) VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE SET type = EXCLUDED.type, config = EXCLUDED.config`
	if _, err := r.Base.DB.Exec(query, userID, notifType, channelConfig); err != nil {
		return err
	}

	return nil
}

// Delete user notification settings
func (r *Repository) deleteUserNotificationSettings(userID int64) error {
	if _, err := r.Base.DB.Exec("DELETE FROM user_notifications WHERE user_id = $1", userID); err != nil {
		return err
	}

	return nil
}

// Delete notifications where the user does not have a notifications channel
func (r *Repository) deleteOrphanedNotifications() (int64, error) {
	query := `DELETE FROM notification_queue
		     WHERE user_id NOT IN (
    	     SELECT user_id FROM user_notifications);`

	result, err := r.Base.DB.Exec(query)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *Repository) deleteNotificationFromQueue(id int64) (int64, error) {
	result, err := r.Base.DB.Exec("DELETE FROM notification_queue WHERE id = $1", id)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}
