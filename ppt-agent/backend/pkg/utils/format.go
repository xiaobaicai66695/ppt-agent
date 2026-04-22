package utils

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func FormatInput(messages []*schema.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var result string
	for i, msg := range messages {
		result += fmt.Sprintf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
	}
	return result
}

func FormatExecutedSteps(steps []*schema.Message) string {
	if len(steps) == 0 {
		return "No steps executed yet."
	}

	var result string
	for _, step := range steps {
		result += "- " + step.Content + "\n"
	}
	return result
}