package users

import (
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"fmt"
	"log/slog"
)

// Change a password for authenticated user
func ChangePassword(repo *repository.Repository, req model.ChangePasswordRequest, logger *slog.Logger) error {
	// Check current password
	isCurrentPassword, err := comparePassword(repo, req.Username, req.OldPassword, logger)
	if err != nil {
		return err
	}

	if !isCurrentPassword {
		return ErrAuthenticationFailed
	}

	//Validate complexity
	if err := validatePassword(req.Username, req.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	// Hash new password and change
	newHash, err := hashString(req.NewPassword)
	if err != nil {
		return err
	}

	if err = repo.ChangePassword(req.Username, newHash); err != nil {
		return err
	}

	return nil
}
