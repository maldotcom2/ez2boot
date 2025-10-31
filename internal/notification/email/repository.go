package email

func (r *Repository) addOrUpdate(userID int64, cfg string) error {
	query := `INSERT INTO user_notifications (user_id, type, config) VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE SET config = EXCLUDED.config
			WHERE user_notifications.config <> EXCLUDED.config`
	if _, err := r.Base.DB.Exec(query, userID, "email", cfg); err != nil {
		return err
	}

	return nil
}
