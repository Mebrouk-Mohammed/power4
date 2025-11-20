package auth

import (
	"context"
	"errors"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type memoryRepo struct {
	mu     sync.RWMutex
	byName map[string]*memoryUser
	nextID int
}

type memoryUser struct {
	u    User
	hash []byte
}

func NewMemoryRepo() Repository {
	return &memoryRepo{byName: make(map[string]*memoryUser), nextID: 1}
}

func (m *memoryRepo) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username required")
	}
	if _, ok := m.byName[username]; ok {
		return nil, errors.New("username exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id := m.nextID
	m.nextID++
	mu := &memoryUser{u: User{ID: id, Username: username, Email: email}, hash: hash}
	m.byName[username] = mu
	return &mu.u, nil
}

func (m *memoryRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	username = strings.TrimSpace(username)
	if mu, ok := m.byName[username]; ok {
		u := mu.u
		return &u, nil
	}
	return nil, nil
}

func (m *memoryRepo) GetByID(ctx context.Context, id int) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, mu := range m.byName {
		if mu.u.ID == id {
			u := mu.u
			return &u, nil
		}
	}
	return nil, nil
}

func (m *memoryRepo) Authenticate(ctx context.Context, username, password string) (*User, error) {
	m.mu.RLock()
	mu, ok := m.byName[username]
	m.mu.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	if err := bcrypt.CompareHashAndPassword(mu.hash, []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	u := mu.u
	return &u, nil
}

func (m *memoryRepo) Close() error { return nil }

// DeleteUser removes a user by numeric ID from the in-memory store.
func (m *memoryRepo) DeleteUser(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, mu := range m.byName {
		if mu.u.ID == id {
			delete(m.byName, name)
			return nil
		}
	}
	return errors.New("not found")
}

// UpdateAvatar sets the AvatarURL for a user by ID.
func (m *memoryRepo) UpdateAvatar(ctx context.Context, id int, avatarURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, mu := range m.byName {
		if mu.u.ID == id {
			mu.u.AvatarURL = avatarURL
			return nil
		}
	}
	return errors.New("not found")
}
