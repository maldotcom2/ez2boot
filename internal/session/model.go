package session

import "ez2boot/internal/db"

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo *Repository
}
