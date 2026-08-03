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
	"sync"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

// Message 表示对话历史中的单条消息。
type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationSession 保存任务对话历史记录。
type ConversationSession struct {
	mu        sync.RWMutex
	TaskID    string    `json:"task_id"`
	WorkDir   string    `json:"work_dir"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddUserMessage 向对话追加用户消息并持久化。
func (s *ConversationSession) AddUserMessage(content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

// AddAssistantMessage 向对话追加助手消息并持久化。
func (s *ConversationSession) AddAssistantMessage(content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

// GetRecentMessages 返回最后 n 条消息用于上下文注入。
func (s *ConversationSession) GetRecentMessages(n int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n <= 0 || len(s.Messages) == 0 {
		return append([]Message(nil), s.Messages...)
	}
	if n >= len(s.Messages) {
		return append([]Message(nil), s.Messages...)
	}
	return append([]Message(nil), s.Messages[len(s.Messages)-n:]...)
}

// Snapshot returns a race-free copy for API responses.
func (s *ConversationSession) Snapshot() ConversationSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ConversationSession{
		TaskID: s.TaskID, WorkDir: s.WorkDir,
		Messages:  append([]Message(nil), s.Messages...),
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

// LoadSessionFromDB 从数据库加载对话历史记录。
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

// NewSession 为任务创建一个新的对话会话。
func NewSession(taskID, workDir string) *ConversationSession {
	return &ConversationSession{
		TaskID:    taskID,
		WorkDir:   workDir,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SessionManager 管理所有对话会话。
// 提供数据库持久化支持的就内存缓存。
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*ConversationSession // key: taskID
}

// NewSessionManager 创建一个新的会话管理器。
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ConversationSession),
	}
}

// GetOrCreate 返回现有会话或创建新会话。
func (sm *SessionManager) GetOrCreate(taskID, workDir string) *ConversationSession {
	sm.mu.RLock()
	if s, ok := sm.sessions[taskID]; ok {
		sm.mu.RUnlock()
		return s
	}
	sm.mu.RUnlock()

	// Load from database first
	s := LoadSessionFromDB(taskID, workDir)
	sm.mu.Lock()
	if existing, ok := sm.sessions[taskID]; ok {
		sm.mu.Unlock()
		return existing
	}
	sm.sessions[taskID] = s
	sm.mu.Unlock()
	return s
}

// Get 根据 taskID 返回会话，如果未找到则返回 nil。
func (sm *SessionManager) Get(taskID string) *ConversationSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[taskID]
}

// Delete 从内存中删除会话。
func (sm *SessionManager) Delete(taskID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, taskID)
}

// DeleteFromDB 从数据库中删除任务的所有消息。
func DeleteSessionFromDB(taskID string) error {
	return db.DeleteConversationMessages(taskID)
}
