package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

// ReadTextFileTool 读取受控工作区中的指定文本文件。
type ReadTextFileTool struct {
	workspace *Workspace
}

// NewReadTextFileTool 创建使用正式工作区的文本读取工具。
func NewReadTextFileTool(workspace *Workspace) (*ReadTextFileTool, error) {
	if workspace == nil || workspace.root == nil {
		return nil, fmt.Errorf("create read_text_file tool: workspace is nil")
	}

	return &ReadTextFileTool{
		workspace: workspace,
	}, nil
}

type readTextFileArguments struct {
	Path string `json:"path"`
}

var readTextFileParameters = json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "工作区相对路径, 使用 '/' 分隔, 可含多层目录, 例如 internal/config/config.go; 不要使用绝对路径"
    }
  },
  "required": ["path"],
  "additionalProperties": false
}`)

func (r *ReadTextFileTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "read_text_file",
		Description: "读取工作区内的一个受支持文本文件, 返回带行号的纯文本。\n" +
			"path 为工作区相对路径; 绝对路径、目录、符号链接、非普通文件、不支持的扩展名, 以及含 '\\' 或 ':' 的路径都会被拒绝; 不确定文件是否存在时先用 search_text 查找。\n" +
			"首行是摘要, 形如: 内容是否被截断: false | bytes: 1521 | runes: 1521 | lines: 97; 第二行是 \"正文:\", 其后每行是 \"行号 | 原文\", 行号按本次渲染出的行数右对齐, 行尾的一个 '\\r' 会被去掉。\n" +
			"bytes 与 runes 是返回正文(截断后)的原始字节数和码点数, 含 '\\r' 与换行符; lines 是整个文件的行数, 由一次独立的全文扫描得出, 空文件为 0。\n" +
			"正文上限 1000000 字节: 超出时只返回前 1000000 字节并丢弃被切断的残缺字节, 此时 内容是否被截断 为 true, bytes 可能略小于 1000000, 而 lines 仍是全文行数。本工具没有偏移或长度参数, 被截断时改用 search_text 定位, 或只读取确实需要的文件。\n" +
			"若文件中存在超过 10485760 字节的单行, 总行数无法统计: 返回已读到的正文并把 lines 置为 0, 同时给出错误说明。其余失败(路径被拒绝、类型不支持、非普通文件、读取出错)只返回失败原因, 不返回内容。",
		Parameters: readTextFileParameters,
	}
}

func (r *ReadTextFileTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 read_text_file 工具前失败, 因为: %w", err)
	}

	args, err := tool.DecodeObjectArguments[readTextFileArguments](arguments)
	if err != nil {
		return "", fmt.Errorf(
			"解析 parse read_text_file 参数失败, 因为: %w",
			err,
		)
	}

	executeResult, err := r.workspace.ReadTextFile(args.Path)
	result := readTextFileToolResultResponse(executeResult)

	if err != nil {
		return result, fmt.Errorf("使用 read_text_file 工具获取文件 %q 内容失败, 因为: %w", args.Path, err)
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("执行 read_text_file 工具后, 即将返回结果时失败, 因为: %w", err)
	}

	return result, nil
}

var _ tool.Tool = (*ReadTextFileTool)(nil)
