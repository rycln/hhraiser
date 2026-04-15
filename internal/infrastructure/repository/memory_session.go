package repository

import (
	"sync"

	"github.com/rycln/hhraiser/internal/domain"
)

type SessionRepo struct {
	mu      sync.RWMutex
	session *domain.Session
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{}
}

func (r *SessionRepo) Save(s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.session = s
	return nil
}
