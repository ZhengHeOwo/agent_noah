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
      "description": "runtime_config 的 status 字段: 空字符串表示本次不做更改; inactive 表示把运行状态置为非活跃"
    }
  },
  "required": ["status"],
  "additionalProperties": false
}`)

func (u *UpdateRuntimeConfigTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "update_runtime_config",
		Description: "更新 Runtime 的运行配置, 当前只支持把 status 置为 inactive(空字符串表示不做更改)。\n" +
			"置为非活跃后, 运行器下次以 inactive 加载上下文时只把最近 10 个 serial 的活跃上下文转入新的 profile, 其余活跃上下文与 history_context 一并丢弃。\n" +
			"本工具只写 runtime_config 表: 运行状态在运行器启动时读取一次, 本次变更不会改变当前进程已加载的状态。",
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
