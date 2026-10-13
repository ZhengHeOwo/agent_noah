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
      "description": "删除目标的serial, 通常显示在content开头"
    },
	 "id": {
      "type": "integer",
      "description": "删除目标的id, 通常显示在content开头"
    }
  },
  "required": ["serial", "id"],
  "additionalProperties": false
}`)

func (d *DeleteActiveContextBySerialAndIDTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "delete_active_context_by_serial_and_id",
		Description: "根据提供的serial和id删除当前活跃上下文的一个指定message, 删除会在本轮Tool全部返回后直接生效, " +
			"不存在重合的(serial, id)组合, " +
			"序列-serial是一次对话的计数, 通常等于对方的说话次数, 一次对话内产生的所有message的serial相同, " +
			"身份-ID是一个message的独有标识,非程序异常不存在相同id.",
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
