package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/openai/openai-go/v3"
)

func executeToolCall(toolCall openai.ChatCompletionMessageToolCallUnion) (string, error) {
	switch toolCall.Function.Name {
	case "Read":
		return executeRead(toolCall.Function.Arguments)
	case "Write":
		return executeWrite(toolCall.Function.Arguments)
	case "Bash":
		return executeBash(toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown tool %s", toolCall.Function.Name)
	}
}

func executeRead(rawArguments string) (string, error) {
	var args struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

func executeWrite(rawArguments string) (string, error) {
	var args struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	if err := os.WriteFile(args.FilePath, []byte(args.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("File written successfully to %s", args.FilePath), nil
}

func executeBash(rawArguments string) (string, error) {
	var args struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	cmd := exec.Command("bash", "-c", args.Command)

	wd, err := os.Getwd()
	if err == nil {
		cmd.Dir = wd
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Sprintf("%s\nError: %v", string(output), err), nil
		}
		return fmt.Sprintf("Error: %v", err), nil
	}

	return string(output), nil
}