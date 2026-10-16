package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

// WriteTextFileTool 经确认后，在受控工作区中创建或覆盖文本文件。
type WriteTextFileTool struct {
	workspace *Workspace
	confirmer tool.Confirmer
}

// NewWriteTextFileTool 创建使用正式工作区的文本写入工具。
func NewWriteTextFileTool(workspace *Workspace, confirmer tool.Confirmer) (*WriteTextFileTool, error) {
	if workspace == nil || workspace.root == nil {
		return nil, fmt.Errorf("create write_text_file tool: workspace is nil")
	}

	if confirmer == nil {
		return nil, fmt.Errorf(
			"create write_text_file tool: confirmer is nil",
		)
	}

	return &WriteTextFileTool{
		workspace: workspace,
		confirmer: confirmer,
	}, nil
}

type writeTextFileArguments struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Reason  string `json:"reason"`
}

var writeTextFileParameters = json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "目标文件的工作区相对路径, 使用 '/' 分隔, 例如 internal/config/config.go; 不要使用绝对路径、反斜杠、冒号, 也不要使用 '.' 或 '..' 路径段。仅支持文件名恰为 go.mod、go.sum, 或扩展名为 .go、.md、.txt、.json、.yaml、.yml、.toml 的文件, 区分大小写。缺失的父目录会被创建; 目标为符号链接或非普通文件时写入失败。"
    },
    "content": {
      "type": "string",
      "description": "文件最终应保存的完整正文, 不是补丁、差异、追加片段或修改指令; 已有文件会被整体覆盖, 需要保留的原内容必须包含在此字符串中。不得为空或仅含空白, 且不得超过 1000000000 字节和 1000000000 个 Unicode 码点。不要包含 read_text_file 附加的摘要、行号、' | ' 分隔符或用于展示代码的围栏。"
    },
    "reason": {
      "type": "string",
      "description": "本次修改的原因, 会显示在审批信息中, 简要说明即可"
    }
  },
  "required": ["path", "content"],
  "additionalProperties": false
}`)

func (t *WriteTextFileTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "write_text_file",
		Description: "经审批后, 在工作区中创建或整体覆盖一个受支持的文本文件。\n" +
			"审批未通过时不产生任何更改, 结果文本会说明写入成功或被拒绝。",
		Parameters: writeTextFileParameters,
	}
}

func (t *WriteTextFileTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 write_text_file 工具前中断, 因为: %w", err)
	}

	args, err := tool.DecodeObjectArguments[writeTextFileArguments](arguments)
	if err != nil {
		return "", fmt.Errorf(
			"解析 write_text_file 参数失败, 因为: %w(未产生审批)",
			err,
		)
	}

	toolPath, err := t.workspace.validateTextFileWrite(args.Path, args.Content)
	if err != nil {
		return "", fmt.Errorf(
			"write_text_file 参数未通过执行前校验: %w(未产生审批)",
			err,
		)
	}

	confirmationDetails := fmt.Sprintf(
		"内容长度: %d 字节\n申请原因:\n%s",
		len(args.Content),
		args.Reason,
	)

	confirmed, err := t.confirmer.Confirm(
		ctx,
		tool.ConfirmationRequest{
			Action:  "write_text_file",
			Summary: fmt.Sprintf("创建或覆盖文件 %s", toolPath),
			Details: confirmationDetails,
		},
	)

	if err != nil {
		return "", fmt.Errorf("授权程序中断, 因为: %w", err)
	}

	if !confirmed {
		return writeTextFileToolResultResponse(&writeTextFileResult{
			toolPath: toolPath,
		}), nil
	}

	result, err := t.workspace.WriteTextFile(toolPath, args.Content)
	if err != nil {
		return "", fmt.Errorf(
			"写入 %q 失败, 因为: %w",
			toolPath,
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 write_text_file 工具后, 即将返回结果时失败, 因为: %w(文件操作已完成)", err)
	}

	return writeTextFileToolResultResponse(result), nil
}

var _ tool.Tool = (*WriteTextFileTool)(nil)
