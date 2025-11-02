package session

import (
	"errors"
	"ez2boot/internal/notification"
	"ez2boot/internal/shared"
	"fmt"
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

// High level for processing server sessions in each state - called by go routine worker
func (s *Service) ProcessServerSessions() {
	// Ready-for-use sessions
	if err := s.processReadyServerSessions(); err != nil {
		s.Logger.Error("Error while processing ready server sessions", "error", err)
	}

	// Aging sessions
	if err := s.processAgingServerSessions(); err != nil {
		s.Logger.Error("Error while processing aging server sessions", "error", err)
	}

	// Expired sessions
	if err := s.processExpiredServerSessions(); err != nil {
		s.Logger.Error("Error while processing expired server sessions", "error", err)
	}

	// Terminated sessions
	if err := s.processTerminatedServerSessions(); err != nil {
		s.Logger.Error("Error while processing terminated server sessions", "error", err)
	}

	// Finalised sessions
	if err := s.processFinalisedServerSessions(); err != nil {
		s.Logger.Error("Error while processing terminated server sessions", "error", err)
	}
}

// Server sessions which are ready for use
func (s *Service) processReadyServerSessions() error {
	sessionsForUse, err := s.Repo.findPendingOnServerSessions()
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

			if err = s.Repo.setOnNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
				s.Logger.Error("Failed up set flag for session notified on", "error", err)
				tx.Rollback()
				continue
			}

			tx.Commit()
		}
	}

	return nil
}

// Process server sessions which will expire soon and user not yet notified
func (s *Service) processAgingServerSessions() error {
	agingSessions, err := s.Repo.findAgingServerSessions()
	if err != nil {
		return err
	}

	if len(agingSessions) == 0 {
		s.Logger.Debug("No aging server sessions")
		return nil
	}

	s.Logger.Debug("Found aging sessions", "count", len(agingSessions))

	// Queue notification for each and set flag
	for _, session := range agingSessions {
		n := notification.NewNotification{
			UserID: session.UserID,
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

	return nil
}

// Process expired server session which haven't been processed yet
func (s *Service) processExpiredServerSessions() error {
	expiredSessions, err := s.Repo.findExpiredServerSessions()
	if err != nil {
		return err
	}

	if len(expiredSessions) == 0 {
		s.Logger.Debug("No expired server sessions")
		return nil
	}

	s.Logger.Debug("Found expired sessions", "count", len(expiredSessions))

	// TODO Add expired notification
	for _, session := range expiredSessions {
		n := notification.NewNotification{
			UserID: session.UserID,
			Msg:    fmt.Sprintf("Your server session has expired for Server Group %s. Servers will power off", session.ServerGroup),
			Title:  fmt.Sprintf("Session expired for server group %s.", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for expired sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		defer tx.Rollback()

		if err := s.NotificationService.QueueNotification(tx, n); err != nil {
			s.Logger.Error("Failed to queue expired session notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		if err := s.Repo.endServerSession(tx, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to end expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		tx.Commit()
	}

	return nil
}

// Process sessions which have been marked for cleanup and users not yet notified
func (s *Service) processTerminatedServerSessions() error {
	terminatedSessions, err := s.Repo.findTerminatedServerSessions()
	if err != nil {
		s.Logger.Error("Error while finding terminated server sessions", "error", err)
	}

	if len(terminatedSessions) == 0 {
		s.Logger.Debug("No terminated server sessions")
		return nil
	}

	for _, session := range terminatedSessions {
		n := notification.NewNotification{
			UserID: session.UserID,
			Msg:    fmt.Sprintf("Your session has ended normally for Server Group %s. Servers are now off", session.ServerGroup),
			Title:  fmt.Sprintf("Session ended for server group %s.", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for terminated sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		defer tx.Rollback()

		if err := s.NotificationService.QueueNotification(tx, n); err != nil {
			s.Logger.Error("Failed to queue aging session notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		if err := s.Repo.setOffNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		tx.Commit()

	}

	return nil
}

// Process sessions which are marked for cleanup and users have been notified server off state
func (s *Service) processFinalisedServerSessions() error {
	sessionsForFinalise, err := s.Repo.findFinalisedServerSessions()
	if err != nil {
		s.Logger.Error("Error while finding sessions for cleanup", "error", err)
	}

	if len(sessionsForFinalise) == 0 {
		s.Logger.Debug("No sessions for cleanup")
	} else {
		s.Repo.cleanupServerSessions(sessionsForFinalise)
	}

	return nil
}
