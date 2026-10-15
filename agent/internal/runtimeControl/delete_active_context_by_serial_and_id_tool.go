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
      "description": "删除目标所属对话的serial, 取自该message内容开头的前缀行(形如 Prefix: | serial 3 | id 47 | ...)。同一次对话内产生的所有message共用同一个serial"
    },
	 "id": {
      "type": "integer",
      "description": "删除目标message的id, 取自该message内容开头的前缀行。id在活跃上下文内唯一, 不存在重复的(serial, id)组合"
    }
  },
  "required": ["serial", "id"],
  "additionalProperties": false
}`)

func (d *DeleteActiveContextBySerialAndIDTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "delete_active_context_by_serial_and_id",
		Description: "从活跃上下文中删除指定的一条message, 用于上下文精炼。serial与id取自该message内容开头的前缀行。" +
			"一次调用删除一条; 同一轮内可以发起多条调用, 全部删除在本轮Tool返回完毕后一并生效。" +
			"删除只作用于活跃上下文, 不改变history_context, 历史底账不受影响。",
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
		return result, fmt.Errorf("执行 delete_active_context_by_serial_and_id 工具后, 即将返回结果时失败, 因为: %w", err)
	}

	return result, nil
}

var _ tool.Tool = (*DeleteActiveContextBySerialAndIDTool)(nil)
