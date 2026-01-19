package audit

import (
	"database/sql"
	"time"
)

func (r *Repository) Log(e Event) error {
	query := "INSERT INTO audit_log (actor_user_id, actor_email, target_user_id, target_email, action, resource, success, reason, time_stamp ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	if _, err := r.Base.DB.Exec(query, e.ActorUserID, e.ActorEmail, e.TargetUserID, e.TargetEmail, e.Action, e.Resource, e.Success, e.Reason, time.Now().Unix()); err != nil {
		return err
	}

	return nil
}

func (r *Repository) LogTx(tx *sql.Tx, e Event) error {
	query := "INSERT INTO audit_log (actor_user_id, actor_email, target_user_id, target_email, action, resource, success, reason, time_stamp ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	if _, err := tx.Exec(query, e.ActorUserID, e.ActorEmail, e.TargetUserID, e.TargetEmail, e.Action, e.Resource, e.Success, e.Reason, time.Now().Unix()); err != nil {
		return err
	}

	return nil
}
