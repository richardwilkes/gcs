/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package websettings

import (
	"context"
	"io/fs"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
)

// Minimums and defaults for web server settings.
var (
	DefaultShutdownGracePeriod = fxp.Ten
	DefaultReadTimeout         = fxp.Ten
	DefaultWriteTimeout        = fxp.Thirty
	DefaultIdleTimeout         = fxp.Sixty
	MinimumTimeout             = fxp.One
	MaximumTimeout             = fxp.SixHundred
	DefaultAddress             = "localhost:8422"
)

// Server holds the settings for the embedded web server.
type Server struct {
	Enabled             bool    `json:"enabled"`
	Address             string  `json:"address,omitempty"`
	CertFile            string  `json:"cert_file,omitempty"`
	KeyFile             string  `json:"key_file,omitempty"`
	ShutdownGracePeriod fxp.Int `json:"shutdown_grace_period,omitempty"`
	ReadTimeout         fxp.Int `json:"read_timeout,omitempty"`
	WriteTimeout        fxp.Int `json:"write_timeout,omitempty"`
	IdleTimeout         fxp.Int `json:"idle_timeout,omitempty"`
}

type wrapper struct {
	Server
	Users    map[string]*User       `json:"users,omitempty"`
	Sessions map[uuid.UUID]*Session `json:"sessions,omitempty"`
}

// Settings holds the settings for the embedded web server.
type Settings struct {
	Server
	lock     sync.RWMutex
	users    map[string]*User
	sessions map[uuid.UUID]*Session
}

// Default returns the default settings.
func Default() *Settings {
	return &Settings{
		Server: Server{
			Address:             DefaultAddress,
			ShutdownGracePeriod: DefaultShutdownGracePeriod,
			ReadTimeout:         DefaultReadTimeout,
			WriteTimeout:        DefaultWriteTimeout,
			IdleTimeout:         DefaultIdleTimeout,
		},
		users:    make(map[string]*User),
		sessions: make(map[uuid.UUID]*Session),
	}
}

// NewSettingsFromFile loads new settings from a file.
func NewSettingsFromFile(fileSystem fs.FS, filePath string) (*Settings, error) {
	var settings Settings
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// Save writes the settings to the file as JSON.
func (s *Settings) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, s)
}

// Valid returns true if the settings are valid.
func (s *Settings) Valid() bool {
	if s.Address == "" {
		return false
	}
	if _, _, err := net.SplitHostPort(s.Address); err != nil {
		return false
	}
	if s.ShutdownGracePeriod < MinimumTimeout || s.ShutdownGracePeriod > MaximumTimeout {
		return false
	}
	if s.ReadTimeout < MinimumTimeout || s.ReadTimeout > MaximumTimeout {
		return false
	}
	if s.WriteTimeout < MinimumTimeout || s.WriteTimeout > MaximumTimeout {
		return false
	}
	if s.IdleTimeout < MinimumTimeout || s.IdleTimeout > MaximumTimeout {
		return false
	}
	return true
}

// Validate the settings.
func (s *Settings) Validate() {
	s.Address = strings.TrimSpace(s.Address)
	if s.Address == "" {
		s.Address = DefaultAddress
	}
	s.ShutdownGracePeriod = min(max(s.ShutdownGracePeriod, MinimumTimeout), MaximumTimeout)
	s.ReadTimeout = min(max(s.ReadTimeout, MinimumTimeout), MaximumTimeout)
	s.WriteTimeout = min(max(s.WriteTimeout, MinimumTimeout), MaximumTimeout)
	s.IdleTimeout = min(max(s.IdleTimeout, MinimumTimeout), MaximumTimeout)
	s.lock.Lock()
	if s.users == nil {
		s.users = make(map[string]*User)
	}
	if s.sessions == nil {
		s.sessions = make(map[uuid.UUID]*Session)
	}
	s.lock.Unlock()
}

// CopyFrom copies the settings from the other Settings to this Settings object. If 'other' is nil, the default settings
// are used.
func (s *Settings) CopyFrom(other *Settings) {
	if other == nil {
		other = Default()
	}
	s.Server = other.Server
	other.lock.RLock()
	users := make(map[string]*User, len(other.users))
	for key, user := range other.users {
		users[key] = user.Clone()
	}
	sessions := make(map[uuid.UUID]*Session, len(other.sessions))
	for id, session := range other.sessions {
		sessions[id] = session.Clone()
	}
	other.lock.RUnlock()
	s.lock.Lock()
	s.users = users
	s.sessions = other.sessions
	s.lock.Unlock()
}

// MarshalJSON implements json.Marshaler.
func (s *Settings) MarshalJSON() ([]byte, error) {
	s.PruneSessions()
	s.lock.RLock()
	defer s.lock.RUnlock()
	return json.Marshal(&wrapper{
		Server:   s.Server,
		Users:    s.users,
		Sessions: s.sessions,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Settings) UnmarshalJSON(data []byte) error {
	s.lock.Lock()
	defer s.Validate()
	defer s.lock.Unlock()
	var w wrapper
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	s.Server = w.Server
	s.users = w.Users
	s.sessions = w.Sessions
	s.pruneSessions()
	return nil
}

// PruneSessions removes expired sessions.
func (s *Settings) PruneSessions() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pruneSessions()
}

func (s *Settings) pruneSessions() {
	var sessionsToDelete []uuid.UUID
	for id, session := range s.sessions {
		if session.Expired() {
			sessionsToDelete = append(sessionsToDelete, id)
		}
	}
	for _, id := range sessionsToDelete {
		delete(s.sessions, id)
	}
}

// LookupUserNameAndPassword looks up a user.
func (s *Settings) LookupUserNameAndPassword(name string) (actualName, hashedPassword string, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var user *User
	if user, ok = s.users[userNameToKey(name)]; !ok {
		return "", "", false
	}
	return user.Name, user.HashedPassword, true
}

// CreateUser creates a user. Returns true on success, false if a user by that name already exists.
func (s *Settings) CreateUser(name, password string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	key := userNameToKey(name)
	if _, exists := s.users[key]; exists {
		return false
	}
	s.users[key] = &User{
		Name:           name,
		HashedPassword: HashPassword(password),
		AccessList:     nil,
	}
	return true
}

// RemoveUser removes a user.
func (s *Settings) RemoveUser(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	key := userNameToKey(name)
	delete(s.users, key)
	var keysToDelete []uuid.UUID
	for id, session := range s.sessions {
		if session.UserKey == key {
			keysToDelete = append(keysToDelete, id)
		}
	}
	for _, id := range keysToDelete {
		delete(s.sessions, id)
	}
}

// RenameUser renames a user. Returns true on success, false if the new name already exists or the user can't be found.
func (s *Settings) RenameUser(oldName, newName string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	oldKey := userNameToKey(oldName)
	user, exists := s.users[oldKey]
	if !exists {
		return false
	}
	user.Name = newName
	newKey := userNameToKey(newName)
	if newKey != oldKey {
		delete(s.users, oldKey)
		s.users[newKey] = user
		for _, session := range s.sessions {
			if session.UserKey == oldKey {
				session.UserKey = newKey
			}
		}
	}
	return true
}

// LookupSession looks up a session, updating its last used time and returning the user's name if found.
func (s *Settings) LookupSession(id uuid.UUID) (string, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return "", false
	}
	var user *User
	if user, ok = s.users[session.UserKey]; !ok || session.Expired() {
		delete(s.sessions, id)
		return "", false
	}
	session.LastUsed = time.Now()
	return user.Name, true
}

// CreateSession creates a session.
func (s *Settings) CreateSession(userName string) uuid.UUID {
	s.lock.Lock()
	defer s.lock.Unlock()
	session := &Session{
		ID:       uuid.New(),
		UserKey:  userNameToKey(userName),
		Issued:   time.Now(),
		LastUsed: time.Now(),
	}
	s.sessions[session.ID] = session
	return session.ID
}

// RemoveSession removes a session.
func (s *Settings) RemoveSession(id uuid.UUID) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.sessions, id)
}
