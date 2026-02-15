package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (r *Repository) GetAuditEvents(req AuditLogRequest) (AuditLogResponse, error) {
	var args []any
	var conditions []string

	// Components of 'where' clause
	if req.From != 0 {
		args = append(args, req.From)
		conditions = append(conditions, fmt.Sprintf("time_stamp >= $%d", len(args)))
	}

	if req.To != 0 {
		args = append(args, req.To)
		conditions = append(conditions, fmt.Sprintf("time_stamp <= $%d", len(args)))
	}

	if req.ActorEmail != "" {
		args = append(args, req.ActorEmail)
		conditions = append(conditions, fmt.Sprintf("actor_email = $%d", len(args)))
	}

	if req.TargetEmail != "" {
		args = append(args, req.TargetEmail)
		conditions = append(conditions, fmt.Sprintf("target_email = $%d", len(args)))
	}

	if req.Action != "" {
		args = append(args, req.Action)
		conditions = append(conditions, fmt.Sprintf("action = $%d", len(args)))
	}

	if req.Resource != "" {
		args = append(args, req.Resource)
		conditions = append(conditions, fmt.Sprintf("resource = $%d", len(args)))
	}

	if req.Success != nil {
		args = append(args, *req.Success)
		conditions = append(conditions, fmt.Sprintf("success = $%d", len(args)))
	}

	if req.Reason != "" {
		args = append(args, req.Reason)
		conditions = append(conditions, fmt.Sprintf("reason = $%d", len(args)))
	}

	if req.Metadata != "" {
		args = append(args, "%"+req.Metadata+"%")
		conditions = append(conditions, fmt.Sprintf("metadata LIKE $%d", len(args)))
	}

	if req.Before != 0 {
		args = append(args, req.Before)
		conditions = append(conditions, fmt.Sprintf("time_stamp < $%d", len(args)))
	}

	// Build query
	var where string
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT actor_user_id, actor_email, target_user_id, target_email, action, resource, success, reason, metadata, time_stamp
						FROM audit_log %s
						ORDER BY time_stamp DESC
						LIMIT $%d`, where, len(args)+1)

	args = append(args, req.Limit)

	// Cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute
	rows, err := r.Base.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return AuditLogResponse{}, err
	}

	defer rows.Close()

	events := []AuditLogEvent{}
	var lastCursor int64

	for rows.Next() {
		var metadata sql.NullString
		var a AuditLogEvent
		err = rows.Scan(&a.ActorUserID, &a.ActorEmail, &a.TargetUserID, &a.TargetEmail, &a.Action, &a.Resource, &a.Success, &a.Reason, &metadata, &a.TimeStamp)
		if err != nil {
			return AuditLogResponse{}, err
		}

		a.Metadata = make(map[string]any)
		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &a.Metadata); err != nil {
				return AuditLogResponse{}, err
			}
		}

		events = append(events, a)
		lastCursor = a.TimeStamp
	}

	// Only provide next cursor when limit is reached
	var nextCursor *int64
	if len(events) == req.Limit {
		nextCursor = &lastCursor
	}

	return AuditLogResponse{
		Events:     events,
		NextCursor: nextCursor,
	}, nil
}

func (r *Repository) Log(e Event) error {
	var metadataJSON *string
	if e.Metadata != nil {
		b, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}
		s := string(b)
		metadataJSON = &s
	}

	query := "INSERT INTO audit_log (actor_user_id, actor_email, target_user_id, target_email, action, resource, success, reason, metadata, time_stamp ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

	if _, err := r.Base.DB.Exec(query, e.ActorUserID, e.ActorEmail, e.TargetUserID, e.TargetEmail, e.Action, e.Resource, e.Success, e.Reason, metadataJSON, time.Now().Unix()); err != nil {
		return err
	}

	return nil
}

func (r *Repository) LogTx(tx *sql.Tx, e Event) error {
	var metadataJSON *string
	if e.Metadata != nil {
		b, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}
		s := string(b)
		metadataJSON = &s
	}

	query := "INSERT INTO audit_log (actor_user_id, actor_email, target_user_id, target_email, action, resource, success, reason, metadata, time_stamp ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

	if _, err := tx.Exec(query, e.ActorUserID, e.ActorEmail, e.TargetUserID, e.TargetEmail, e.Action, e.Resource, e.Success, e.Reason, metadataJSON, time.Now().Unix()); err != nil {
		return err
	}

	return nil
}
