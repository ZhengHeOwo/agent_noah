package runtimeControl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository/sqlite"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

type UpdateRuntimeConfigTool struct {
	sqlite *sqlite.Sqlite
}

func NewUpdateRuntimeConfigTool(sqlite *sqlite.Sqlite) (*UpdateRuntimeConfigTool, error) {
	return &UpdateRuntimeConfigTool{
		sqlite: sqlite,
	}, nil
}

type updateRuntimeConfigArguments struct {
	Status string `json:"status"`
}

var updateRuntimeConfigParameters = json.RawMessage(`{
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["", "inactive"],
      "description": "更改runtime_config的status字段: 空字符串表示本次不做任何更改; inactive表示将运行状态设为非活跃, 该变更在本轮RunTurn结束后生效, 之后程序会分离当前history_context用于事件分析, 并在下次进入active时启用全新的history_context表。恢复active由程序自动执行"
    }
  },
  "required": ["status"],
  "additionalProperties": false
}`)

func (u *UpdateRuntimeConfigTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "update_runtime_config",
		Description: "更新Runtime的运行配置。当前仅用于把运行状态置为inactive: 置为非活跃后, 本轮RunTurn结束时程序会分离当前history_context用于事件分析," +
			"并在下次进入active时启用全新的history_context表。适合在一段工作收尾、需要整体换代上下文时使用。",
		Parameters: updateRuntimeConfigParameters,
	}
}

func (u *UpdateRuntimeConfigTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 update_runtime_config 工具前中断, 因为: %w", err)
	}

	args, err := tool.DecodeObjectArguments[updateRuntimeConfigArguments](arguments)
	if err != nil {
		return "", fmt.Errorf(
			"解析 update_runtime_config 参数失败, 因为: %w", err)
	}

	tx, err := u.sqlite.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("update_runtime_config_Tool事务创建失败 %w", err)
	}
	defer tx.Rollback()

	var gatherResult string
	switch args.Status {
	case "":
		gatherResult += fmt.Sprintf("status 无操作 | ")

	case "inactive":
		if err := u.sqlite.UpdateRuntimeConfig(tx, "status", "inactive"); err != nil {
			return "", err
		}
		gatherResult += fmt.Sprintf("成功将 status 更新为 inactive | ")

	default:
		return "", fmt.Errorf("status只允许空字符串或 inactive")
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 update_runtime_config 工具后, 即将返回结果时失败, 因为: %w(操作已执行)", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("update_runtime_config_Tool 出现事务提交异常 %w", err)
	}

	return strings.TrimSpace(gatherResult), nil
}

var _ tool.Tool = (*UpdateRuntimeConfigTool)(nil)
