package model

import (
	"context"
	"encoding/json"
)

// Role 表示消息发送方的角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolDefinition 描述模型可以选择和调用的工具。
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall 描述模型发起的一次工具调用。
type ToolCall struct {
	// ID 必须原样写入对应工具结果消息的ToolCallID。
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"argumnets"`
}

// Prefix 消息的固定前缀.
type Prefix struct {
	Serial         int64  `json:"serial"`
	Id             int64  `json:"id"`
	OccurrenceTime string `json:"occurrence_time"`
	MaxModelSteps  int64  `json:"maxModelSteps"`
	NowSteps       int64  `json:"nowSteps"`
}

// Message 表示与模型供应商无关的对话消息。
type Message struct {
	Role             Role       `json:"role"`
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content"`
	ToolCalls        []ToolCall `json:"tool_calls"`
	ToolCallID       string     `json:"tool_call_id"`
	Prefix           Prefix     `json:"prefix"`
}

// Request 表示一次模型请求。
type Request struct {
	Model    string
	Messages []Message
	Tools    []ToolDefinition
}

// Response 表示一次模型响应。
type Response struct {
	Message      Message
	FinishReason string
}

// Model 定义模型客户端需要实现的能力。
type Model interface {
	Complete(ctx context.Context, request Request) (Response, error)
}
