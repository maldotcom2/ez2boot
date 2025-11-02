package session

import (
	"database/sql"
	"errors"
	"ez2boot/internal/notification"
	"ez2boot/internal/shared"
	"fmt"
	"time"
)

func (s *Service) getServerSessions() ([]ServerSession, error) {
	sessions, err := s.Repo.getServerSessions()
	if err != nil {
		return []ServerSession{}, err
	}

	return sessions, nil
}

func (s *Service) newServerSession(session ServerSession) (ServerSession, error) {
	if session.ServerGroup == "" || session.Duration == "" {
		return ServerSession{}, errors.New("Server_group and duration required") //TODO Sentinel errors
	}

	// Get email for user
	email, err := s.UserService.FindEmailFromUserID(session.UserID)
	if err != nil {
		return ServerSession{}, err
	}

	session.Email = email

	session, err = s.Repo.newServerSession(session)
	if err != nil {
		return ServerSession{}, err
	}

	return session, nil
}

func (s *Service) updateServerSession(session ServerSession) (ServerSession, error) {
	if session.ServerGroup == "" || session.Duration == "" {
		return ServerSession{}, errors.New("email, server_group, duration is required")
	}

	updated, updatedSession, err := s.Repo.updateServerSession(session)
	if err != nil {
		return ServerSession{}, err
	}

	if !updated {
		return ServerSession{}, shared.ErrSessionNotFound
	}

	return updatedSession, nil
}

// High level for processing server sessions in various states - called by go routine worker
func (s *Service) ProcessServerSessions() error {
	// Expired or aging sessions
	if err := s.processExpiredOrAgingServerSessions(); err != nil {
		return err
	}

	// Terminated sessions
	if err := s.processTerminatedServerSessions(); err != nil {
		return err
	}

	// Ready-for-use sessions
	if err := s.processReadyServerSessions(); err != nil {
		return err
	}

	return nil
}

func (s *Service) processExpiredOrAgingServerSessions() error {
	expiredSessions, agingSessions, err := s.findExpiredOrAgingServerSessions()
	if err != nil {
		s.Logger.Error("Error when trying to find aging or expired server sessions", "error", err)
	}

	if len(expiredSessions) == 0 {
		s.Logger.Debug("No expired server sessions")
	} else {
		s.processExpiredServerSessions(expiredSessions)
	}

	if len(agingSessions) == 0 {
		s.Logger.Debug("No server sessions nearing expiry")
	} else {
		s.processAgingServerSessions(agingSessions) // ??
	}

	return nil
}

func (s *Service) processTerminatedServerSessions() error {
	sessionsForCleanup, err := s.findServerSessionsForAction(1, 1, 1, "off")
	if err != nil {
		s.Logger.Error("Error while finding sessions for cleanup", "error", err)
	}

	if len(sessionsForCleanup) == 0 {
		s.Logger.Debug("No sessions for cleanup")
	} else {
		s.Repo.cleanupServerSessions(sessionsForCleanup)
	}

	return nil
}

// Server sessions which are ready for use
func (s *Service) processReadyServerSessions() error {
	sessionsForUse, err := s.findServerSessionsForAction(0, 0, 0, "on")
	if err != nil {
		s.Logger.Error("Error while finding sessions ready for use", "error", err)
	}

	if len(sessionsForUse) == 0 {
		s.Logger.Debug("No new sessions ready for use")
	} else {
		s.Logger.Debug("New sessions ready for use")

		// Queue notification and set notified flag for each
		for _, session := range sessionsForUse {
			n := notification.NewNotification{
				UserID: session.UserID,
				Msg:    fmt.Sprintf("Servers are online and ready for use: Server Group: %s", session.ServerGroup),
				Title:  fmt.Sprintf("Server Group %s online", session.ServerGroup),
			}

			tx, err := s.Repo.Base.DB.Begin()
			if err != nil {
				s.Logger.Error("Failed to create transaction for processing ready session", "email", session.Email, "error", err)
				continue
			}

			defer tx.Rollback()

			if err := s.NotificationService.QueueNotification(tx, n); err != nil {
				s.Logger.Error("Failed to queue sesion ready notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
				tx.Rollback()
				continue
			}

			if err = s.SetOnNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
				s.Logger.Error("Failed up set flag for session notified on", "error", err)
				tx.Rollback()
				continue
			}

			tx.Commit()
		}
	}

	return nil
}

// Find expired sessions
func (s *Service) findExpiredOrAgingServerSessions() ([]ServerSession, []ServerSession, error) {
	currentSessions, err := s.Repo.getServerSessions()
	if err != nil {
		return nil, nil, err
	}

	var expiredSessions []ServerSession
	var agingSessions []ServerSession
	now := time.Now().UTC()
	warningWindow := now.Add(15 * time.Minute) //TODO make adjustable

	for _, session := range currentSessions {
		if session.Expiry.Before(now) {
			expiredSessions = append(expiredSessions, session)
		} else if session.Expiry.Before(warningWindow) {
			agingSessions = append(agingSessions, session)
		}
	}

	return expiredSessions, agingSessions, nil
}

func (s *Service) processExpiredServerSessions(expiredSessions []ServerSession) {
	s.Logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		if err := s.Repo.endServerSession(session.ServerGroup); err != nil {
			s.Logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) processAgingServerSessions(agingSessions []ServerSession) {
	s.Logger.Debug("Found aging sessions", "count", len(agingSessions))

	// Queue notification for each and set flag
	for _, session := range agingSessions {
		n := notification.NewNotification{
			UserID: session.UserID, // Not working
			Msg:    fmt.Sprintf("Session is expiring soon for Server Group %s and can be extended", session.ServerGroup),
			Title:  fmt.Sprintf("Session for Server Group %s expiring", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for aging sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		defer tx.Rollback()

		if err := s.NotificationService.QueueNotification(tx, n); err != nil {
			s.Logger.Error("Failed to queue aging session notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		if err := s.Repo.setWarningNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		tx.Commit()
	}
}

func (s *Service) findServerSessionsForAction(toCleanup int, onNotified int, offNotified int, serverState string) ([]ServerSession, error) {
	sessions, err := s.Repo.findServerSessionsForAction(toCleanup, onNotified, offNotified, serverState)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *Service) SetWarningNotifiedFlag(tx *sql.Tx, flagValue int, serverGroup string) error {
	if err := s.Repo.setWarningNotifiedFlag(tx, flagValue, serverGroup); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetOnNotifiedFlag(tx *sql.Tx, flagValue int, serverGroup string) error {
	if err := s.Repo.setOnNotifiedFlag(tx, flagValue, serverGroup); err != nil {
		return err
	}

	return nil
}
