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

package utils

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// RuntimeStatusChatModel injects a code-maintained status bar at the end of
// each model call. The injected message is appended, so it does not rewrite the
// stable system/tool prefix and remains KV-cache friendly.
type RuntimeStatusChatModel struct {
	inner model.ToolCallingChatModel
	meta  *RuntimeMeta
}

func NewRuntimeStatusChatModel(inner model.ToolCallingChatModel, meta *RuntimeMeta) model.ToolCallingChatModel {
	if inner == nil || meta == nil {
		return inner
	}
	return &RuntimeStatusChatModel{inner: inner, meta: meta}
}

func (m *RuntimeStatusChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.inner.Generate(ctx, m.withStatus(messages), opts...)
}

func (m *RuntimeStatusChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.inner.Stream(ctx, m.withStatus(messages), opts...)
}

func (m *RuntimeStatusChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	innerWithTools, err := m.inner.WithTools(tools)
	if err != nil {
		return nil, err
	}
	return &RuntimeStatusChatModel{inner: innerWithTools, meta: m.meta}, nil
}

func (m *RuntimeStatusChatModel) withStatus(messages []*schema.Message) []*schema.Message {
	if m == nil || m.meta == nil {
		return messages
	}
	status := m.meta.StatusBar()
	if status == "" {
		return messages
	}
	next := make([]*schema.Message, 0, len(messages)+1)
	next = append(next, messages...)
	next = append(next, schema.UserMessage(status))
	return next
}
