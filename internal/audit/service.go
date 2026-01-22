package audit

import "database/sql"

func (s *Service) GetAuditEvents(req AuditLogRequest) (AuditLogResponse, error) {
	events, err := s.Repo.GetAuditEvents(req)
	if err != nil {
		return AuditLogResponse{}, err
	}

	return events, nil
}

func (s *Service) Log(e Event) {
	if err := s.Repo.Log(e); err != nil {
		s.Logger.Error("Failed to write audit log", "error", err)
	}
}

// For use with embedded logging with loops and transactions
func (s *Service) LogTx(tx *sql.Tx, e Event) {
	if err := s.Repo.LogTx(tx, e); err != nil {
		s.Logger.Error("Failed to write audit log", "error", err)
	}
}
