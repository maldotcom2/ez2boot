package repository

import (
	"ez2boot/internal/models"
	"ez2boot/internal/utils"
	"log/slog"
	"time"
)

// Return all servers from catalogue - names and groups
func (r *Repository) GetServers(logger *slog.Logger) ([]models.Server, error) {
	rows, err := r.DB.Query("SELECT name, server_group FROM servers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servers := []models.Server{}
	for rows.Next() {
		var s models.Server
		err = rows.Scan(&s.Name, &s.ServerGroup)
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}

// Return currently active sessions
func (r *Repository) GetSessions(logger *slog.Logger) ([]models.Session, error) {
	rows, err := r.DB.Query("SELECT email, server_group, expiry FROM sessions")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		var s models.Session
		err = rows.Scan(&s.Email, &s.ServerGroup, &s.Expiry)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Create a new session
func (r *Repository) NewSession(session models.Session, logger *slog.Logger) (models.Session, error) {
	newExpiry, err := utils.GetExpiry(0, session.Duration)
	if err != nil {
		return session, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	_, err = r.DB.Exec("INSERT INTO sessions (token, email, server_group, expiry) VALUES ($1, $2, $3, $4)", session.Token, session.Email, session.ServerGroup, newExpiry)
	if err != nil {
		// TO DO: Add error for non-unique where server group already has a session
		return session, err
	}

	return session, nil
}
