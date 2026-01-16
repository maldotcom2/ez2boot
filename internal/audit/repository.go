package audit

import "time"

func (r *Repository) Log(e Event) error {
	query := "INSERT INTO audit_log (user_id, email, action, resource, result, time_stamp ) VALUES ($1, $2, $3, $4, $5, $6)"

	if _, err := r.Base.DB.Exec(query, e.UserID, e.Email, e.Action, e.Resource, e.Result, time.Now().Unix()); err != nil {
		return err
	}

	return nil
}
