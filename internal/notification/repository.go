package notification

// Find all pending notifications in queue and match to user config
func (r *Repository) findPendingNotifications() ([]Notification, error) {
	query := `SELECT nq.message, nq.title, un.type, un.config
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
		if err := rows.Scan(&n.Msg, &n.Title, &n.Type, &n.Cfg); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}
