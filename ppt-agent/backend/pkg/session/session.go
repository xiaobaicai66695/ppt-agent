/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package session

import (
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

// Message represents a single message in the conversation history.
type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationSession holds the conversation history for a task.
type ConversationSession struct {
	TaskID    string    `json:"task_id"`
	WorkDir   string    `json:"work_dir"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddUserMessage appends a user message to the conversation and persists it.
func (s *ConversationSession) AddUserMessage(content string) error {
	msg := db.ConversationMessage{
		TaskID:    s.TaskID,
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	if err := db.CreateConversationMessage(&msg); err != nil {
		return err
	}
	s.Messages = append(s.Messages, Message{
		Role:      "user",
		Content:   content,
		Timestamp: msg.Timestamp,
	})
	s.UpdatedAt = time.Now()
	return nil
}

// AddAssistantMessage appends an assistant message to the conversation and persists it.
func (s *ConversationSession) AddAssistantMessage(content string) error {
	msg := db.ConversationMessage{
		TaskID:    s.TaskID,
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}
	if err := db.CreateConversationMessage(&msg); err != nil {
		return err
	}
	s.Messages = append(s.Messages, Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: msg.Timestamp,
	})
	s.UpdatedAt = time.Now()
	return nil
}

// GetRecentMessages returns the last n messages for context injection.
func (s *ConversationSession) GetRecentMessages(n int) []Message {
	if n <= 0 || len(s.Messages) == 0 {
		return s.Messages
	}
	if n >= len(s.Messages) {
		return s.Messages
	}
	return s.Messages[len(s.Messages)-n:]
}

// LoadSessionFromDB loads conversation history from the database.
func LoadSessionFromDB(taskID, workDir string) *ConversationSession {
	msgs, err := db.ListConversationMessages(taskID)
	if err != nil {
		return &ConversationSession{
			TaskID:    taskID,
			WorkDir:   workDir,
			Messages:  []Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	messages := make([]Message, 0, len(msgs))
	createdAt := time.Now()
	for _, m := range msgs {
		messages = append(messages, Message{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
		if m.Timestamp.Before(createdAt) {
			createdAt = m.Timestamp
		}
	}

	return &ConversationSession{
		TaskID:    taskID,
		WorkDir:   workDir,
		Messages:  messages,
		CreatedAt: createdAt,
		UpdatedAt: time.Now(),
	}
}

// NewSession creates a new conversation session for a task.
func NewSession(taskID, workDir string) *ConversationSession {
	return &ConversationSession{
		TaskID:    taskID,
		WorkDir:   workDir,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SessionManager manages all conversation sessions.
// It provides in-memory caching backed by database persistence.
type SessionManager struct {
	sessions map[string]*ConversationSession // key: taskID
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ConversationSession),
	}
}

// GetOrCreate returns an existing session or creates a new one.
func (sm *SessionManager) GetOrCreate(taskID, workDir string) *ConversationSession {
	if s, ok := sm.sessions[taskID]; ok {
		return s
	}

	// Load from database first
	s := LoadSessionFromDB(taskID, workDir)
	sm.sessions[taskID] = s
	return s
}

// Get returns a session by taskID, or nil if not found.
func (sm *SessionManager) Get(taskID string) *ConversationSession {
	return sm.sessions[taskID]
}

// Delete removes a session from memory.
func (sm *SessionManager) Delete(taskID string) {
	delete(sm.sessions, taskID)
}

// DeleteFromDB removes all messages for a task from the database.
func DeleteSessionFromDB(taskID string) error {
	return db.DeleteConversationMessages(taskID)
}
