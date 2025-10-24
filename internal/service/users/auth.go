package users

import (
	"database/sql"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"time"

	"github.com/alexedwards/argon2id"
)

func ComparePassword(repo *repository.Repository, username string, password string) (bool, error) {
	hash, err := repo.FindHashByUsername(username)
	if err != nil {
		return false, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func GetSessionInfo(repo *repository.Repository, token string) (model.UserSession, error) {
	u, err := repo.FindUserInfoByToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserSession{}, ErrSessionNotFound
		}
		return model.UserSession{}, err
	}

	if u.SessionExpiry < time.Now().Unix() {
		return model.UserSession{}, ErrSessionExpired
	}

	return u, nil
}

func GetBasicAuthInfo(repo *repository.Repository, username string) (int64, error) {
	userID, err := repo.FindBasicAuthUserID(username)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
