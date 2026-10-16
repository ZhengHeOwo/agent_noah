package runtimeControl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository/sqlite"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

type DeleteActiveContextBySerialAndIDTool struct {
	sqlite *sqlite.Sqlite
}

func NewDeleteActiveContextBySerialAndIDTool(sqlite *sqlite.Sqlite) (*DeleteActiveContextBySerialAndIDTool, error) {
	return &DeleteActiveContextBySerialAndIDTool{
		sqlite: sqlite,
	}, nil
}

type deleteActiveContextBySerialAndIDArguments struct {
	Serial int `json:"serial"`
	Id     int `json:"id"`
}

var deleteActiveContextBySerialAndIDParameters = json.RawMessage(`{
  "type": "object",
  "properties": {
    "serial": {
      "type": "integer",
      "description": "目标 message 所属对话的 serial, 取自该 message 内容开头的前缀行(形如 Prefix: | serial 3 | id 47 | ...); 同一次对话内产生的所有 message 共用同一个 serial"
    },
    "id": {
      "type": "integer",
      "description": "目标 message 的 id, 取自同一前缀行; 活跃上下文内 (serial, id) 唯一"
    }
  },
  "required": ["serial", "id"],
  "additionalProperties": false
}`)

func (d *DeleteActiveContextBySerialAndIDTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "delete_active_context_by_serial_and_id",
		Description: "从活跃上下文中删除一条 message, 参数取自该 message 内容开头的前缀行。每条调用只删一条, 立即执行并落库; 同一轮内先提交的删除会影响其后读取到的活跃上下文。\n" +
			"一条带 tool_calls 的 assistant 消息与其全部 tool 结果消息不可拆分, 必须整组删除: 只删组内一部分会使后续请求不合法, 且程序无法自行修复。\n" +
			"删除只作用于活跃上下文, 不改变 history_context。",
		Parameters: deleteActiveContextBySerialAndIDParameters,
	}
}

func (d *DeleteActiveContextBySerialAndIDTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行delete_active_context_by_serial_and_id 工具前中断, 因为: %w", err)
	}

	args, err := tool.DecodeObjectArguments[deleteActiveContextBySerialAndIDArguments](arguments)
	if err != nil {
		return "", fmt.Errorf(
			"解析 delete_active_context_by_serial_and_id 参数失败, 因为: %w", err)
	}

	result, err := d.sqlite.DeleteActiveContextBySerialAndID(args.Serial, args.Id)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("执行delete_active_context_by_serial_and_id 工具后, 即将返回结果时失败, 因为: %w", err)
	}

	return result, nil
}

var _ tool.Tool = (*DeleteActiveContextBySerialAndIDTool)(nil)
