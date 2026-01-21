package session

import (
	"context"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/notification"
	"ez2boot/internal/util"
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

func (s *Service) getServerSessionSummary() ([]ServerSessionSummary, error) {
	summary, err := s.Repo.getServerSessionSummary()
	if err != nil {
		return []ServerSessionSummary{}, err
	}

	return summary, nil
}

func (s *Service) newServerSession(session ServerSessionRequest, ctx context.Context) (time.Time, error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	if err := s.validateServerSession(session); err != nil {
		return time.Time{}, err
	}

	// Get expiry as epoch
	sessionExpiry, err := util.GetExpiryFromDuration(session.Duration)
	if err != nil {
		return time.Time{}, err
	}

	// Add expiry
	session.Expiry = sessionExpiry

	err = s.Repo.newServerSession(session)
	if err != nil {
		return time.Time{}, err
	}

	s.Audit.Log(audit.Event{
		ActorUserID: actorUserID,
		ActorEmail:  actorEmail,
		Action:      "new",
		Resource:    "server session",
		Success:     true,
		Metadata: map[string]any{
			"server_group": session.ServerGroup,
		},
	})

	// Time in time format
	return time.Unix(sessionExpiry, 0).UTC(), nil
}

func (s *Service) updateServerSession(session ServerSessionRequest, ctx context.Context) (time.Time, error) {
	if err := s.validateServerSession(session); err != nil {
		return time.Time{}, err
	}

	// Get new expiry as epoch
	newExpiry, err := util.GetExpiryFromDuration(session.Duration)
	if err != nil {
		return time.Time{}, err
	}

	// Add expiry
	session.Expiry = newExpiry

	err = s.Repo.updateServerSession(session)
	if err != nil {
		return time.Time{}, err
	}

	actorUserID, actorEmail := ctxutil.GetActor(ctx)
	s.Audit.Log(audit.Event{
		ActorUserID: actorUserID,
		ActorEmail:  actorEmail,
		Action:      "update",
		Resource:    "server session",
		Success:     true,
		Metadata: map[string]any{
			"server_group": session.ServerGroup,
		},
	})

	// Time in time format
	return time.Unix(newExpiry, 0).UTC(), nil
}

// High level for processing server sessions in each state - called by go routine worker
func (s *Service) ProcessServerSessions(ctx context.Context) {
	// Ready-for-use sessions
	if err := s.processReadyServerSessions(ctx); err != nil {
		s.Logger.Error("Error while processing ready server sessions", "error", err)
	}

	// Expiring sessions
	if err := s.processExpiringServerSessions(ctx); err != nil {
		s.Logger.Error("Error while processing expiring server sessions", "error", err)
	}

	// Expired sessions
	if err := s.processExpiredServerSessions(ctx); err != nil {
		s.Logger.Error("Error while processing expired server sessions", "error", err)
	}

	// Terminated sessions
	if err := s.processTerminatedServerSessions(ctx); err != nil {
		s.Logger.Error("Error while processing terminated server sessions", "error", err)
	}

	// Finalised sessions
	if err := s.processFinalisedServerSessions(ctx); err != nil {
		s.Logger.Error("Error while processing terminated server sessions", "error", err)
	}
}

// Server sessions which are ready for use
func (s *Service) processReadyServerSessions(ctx context.Context) error {
	sessionsForUse, err := s.Repo.getPendingOnServerSessions()
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
				Msg:    fmt.Sprintf("Servers are online and ready for Server Group: %s", session.ServerGroup),
				Title:  fmt.Sprintf("Session ready: %s", session.ServerGroup),
			}

			tx, err := s.Repo.Base.DB.Begin()
			if err != nil {
				s.Logger.Error("Failed to create transaction for processing ready session", "email", session.Email, "error", err)
				continue
			}

			if err := s.NotificationService.QueueNotification(tx, n); err != nil {
				s.Logger.Error("Failed to queue session ready notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
				tx.Rollback()
				continue
			}

			if err = s.Repo.setOnNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
				s.Logger.Error("Failed up set flag for session notified on", "error", err)
				tx.Rollback()
				continue
			}

			actorUserID, actorEmail := ctxutil.GetActor(ctx)
			s.Audit.LogTx(tx, audit.Event{
				ActorUserID: actorUserID,
				ActorEmail:  actorEmail,
				Action:      "ready",
				Resource:    "server session",
				Success:     true,
				Metadata: map[string]any{
					"server_group": session.ServerGroup,
				},
			})

			tx.Commit()
		}
	}

	return nil
}

// Process server sessions which will expire soon and user not yet notified
func (s *Service) processExpiringServerSessions(ctx context.Context) error {
	expiringSessions, err := s.Repo.getExpiringServerSessions()
	if err != nil {
		return err
	}

	if len(expiringSessions) == 0 {
		s.Logger.Debug("No expiring server sessions")
		return nil
	}

	s.Logger.Debug("Found expiring sessions", "count", len(expiringSessions))

	// Queue notification for each and set flag
	for _, session := range expiringSessions {
		n := notification.NewNotification{
			UserID: session.UserID,
			Msg:    fmt.Sprintf("Your session is expiring soon for Server Group %s and can be extended", session.ServerGroup),
			Title:  fmt.Sprintf("Session expiring: %s", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for expiring sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		if err := s.NotificationService.QueueNotification(tx, n); err != nil {
			s.Logger.Error("Failed to queue expiring session notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		if err := s.Repo.setWarningNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set expiring session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		actorUserID, actorEmail := ctxutil.GetActor(ctx)
		s.Audit.LogTx(tx, audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "expiring",
			Resource:    "server session",
			Success:     true,
			Metadata: map[string]any{
				"server_group": session.ServerGroup,
			},
		})

		tx.Commit()
	}

	return nil
}

// Process expired server session which haven't been processed yet
func (s *Service) processExpiredServerSessions(ctx context.Context) error {
	expiredSessions, err := s.Repo.getExpiredServerSessions()
	if err != nil {
		return err
	}

	if len(expiredSessions) == 0 {
		s.Logger.Debug("No expired server sessions")
		return nil
	}

	s.Logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		n := notification.NewNotification{
			UserID: session.UserID,
			Msg:    fmt.Sprintf("Your session has expired for Server Group %s. Servers will power off", session.ServerGroup),
			Title:  fmt.Sprintf("Session expired: %s", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for expired sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

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

		actorUserID, actorEmail := ctxutil.GetActor(ctx)
		s.Audit.LogTx(tx, audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "expired",
			Resource:    "server session",
			Success:     true,
			Metadata: map[string]any{
				"server_group": session.ServerGroup,
			},
		})

		tx.Commit()
	}

	return nil
}

// Process sessions which have been marked for cleanup and users not yet notified
func (s *Service) processTerminatedServerSessions(ctx context.Context) error {
	terminatedSessions, err := s.Repo.getTerminatedServerSessions()
	if err != nil {
		s.Logger.Error("Error while finding terminated server sessions", "error", err)
	}

	if len(terminatedSessions) == 0 {
		s.Logger.Debug("No terminated server sessions")
		return nil
	}

	for _, session := range terminatedSessions {
		notification := notification.NewNotification{
			UserID: session.UserID,
			Msg:    fmt.Sprintf("Your session has terminated normally for Server Group %s. Servers are now off", session.ServerGroup),
			Title:  fmt.Sprintf("Session terminated: %s", session.ServerGroup),
		}

		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for terminated sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		if err := s.NotificationService.QueueNotification(tx, notification); err != nil {
			s.Logger.Error("Failed to queue terminated session notification", "email", session.Email, "server group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		if err := s.Repo.setOffNotifiedFlag(tx, 1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set session as off notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		actorUserID, actorEmail := ctxutil.GetActor(ctx)
		s.Audit.LogTx(tx, audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "terminated",
			Resource:    "server session",
			Success:     true,
			Metadata: map[string]any{
				"server_group": session.ServerGroup,
			},
		})

		tx.Commit()

	}

	return nil
}

// Process sessions which are marked for cleanup and users have been notified server off state. Restores server group to state ready for new session.
func (s *Service) processFinalisedServerSessions(ctx context.Context) error {
	sessionsForFinalise, err := s.Repo.getFinalisedServerSessions()
	if err != nil {
		s.Logger.Error("Error while finding sessions for cleanup", "error", err)
	}

	if len(sessionsForFinalise) == 0 {
		s.Logger.Debug("No sessions for cleanup")

		return nil
	}

	for _, session := range sessionsForFinalise {
		tx, err := s.Repo.Base.DB.Begin()
		if err != nil {
			s.Logger.Error("Failed to create transaction for finalise sessions", "email", session.Email, "server group", session.ServerGroup, "error", err)
			continue
		}

		if err := s.Repo.cleanupServerSession(tx, session); err != nil {
			s.Logger.Error("Failed to cleanup server session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			tx.Rollback()
			continue
		}

		actorUserID, actorEmail := ctxutil.GetActor(ctx)
		s.Audit.LogTx(tx, audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "finalised",
			Resource:    "server session",
			Success:     true,
			Metadata: map[string]any{
				"server_group": session.ServerGroup,
			},
		})

		tx.Commit()
	}

	return nil
}
