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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func FormatInput(input []adk.Message) string {
	var sb strings.Builder
	for _, msg := range input {
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func ToJSONString(v interface{}) string {
	str, _ := json.Marshal(v)
	return string(str)
}

func PtrOf[T any](v T) *T {
	return &v
}

func GetSessionValue[T any](ctx context.Context, key string) (T, bool) {
	v, ok := adk.GetSessionValue(ctx, key)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return t, true
}

func FormatExecutedSteps(in []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for idx, m := range in {
		_, _ = fmt.Fprintf(&sb, "## %d. Step: %v\n  Result: %v\n\n", idx+1, m.Step, m.Result)
	}
	return sb.String()
}

func FormatExecutedStepsStr(executedSteps string) string {
	if executedSteps == "" {
		return "暂无"
	}
	return executedSteps
}
