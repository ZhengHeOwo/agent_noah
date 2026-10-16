package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

// ListTextFilesTool 列出受控工作区中的可读取文本文件。
type ListTextFilesTool struct {
	workspace *Workspace
}

// NewListTextFilesTool 创建使用正式工作区的文件列表工具。
func NewListTextFilesTool(workspace *Workspace) (*ListTextFilesTool, error) {
	if workspace == nil || workspace.root == nil {
		return nil, fmt.Errorf("create list_text_files tool: workspace is nil")
	}

	return &ListTextFilesTool{
		workspace: workspace,
	}, nil
}

var listTextFilesParameters = json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`)

func (l *ListTextFilesTool) Definition() model.ToolDefinition {
	return model.ToolDefinition{
		Name: "list_text_files",
		Description: "递归列出工作区内可读取的文本文件: 文件名恰为 go.mod、go.sum, 或扩展名恰为 .go、.md、.txt、.json、.yaml、.yml、.toml, 只匹配小写; 目录、符号链接与其他非普通文件会被跳过。\n" +
			"不需要参数。返回纯文本: 第一行是 列表是否被截断: false |, 第二行是 列表结果:, 第三行是 Go 字符串字面量形式的路径列表, 形如 [\"go.mod\" \"main.go\"], 路径为工作区相对路径、以 '/' 分隔; 顺序为目录深度优先遍历, 同一目录内按名称排序。\n" +
			"遍历到的条目超过 2000 个, 或已列出的文件达到 1500 个时标记为截断, 缺少遍历顺序靠后的部分, 且没有分页或续读方式, 应改用更精确的工具定位文件。\n" +
			"遍历失败(如某个条目无法访问)时直接报错, 不返回此前已收集的路径。",
		Parameters: listTextFilesParameters,
	}
}

type listTextFilesResponse struct {
	Paths     []string `json:"paths"`
	Truncated bool     `json:"truncated"`
}

func (l *ListTextFilesTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("执行 list_text_files 工具前失败, 因为: %w", err)
	}

	_, err := tool.DecodeObjectArguments[struct{}](arguments)
	if err != nil {
		return "", fmt.Errorf(
			"解析 list_text_files 参数失败, 因为: %w",
			err,
		)
	}

	executeResult, err := l.workspace.ListTextFiles(ctx)
	result := listTextFilesToolResultResponse(executeResult)

	if err != nil {
		return result, fmt.Errorf("使用 list_text_files 获取文件列表失败, 因为: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("执行 list_text_files 工具后, 即将返回结果时失败, 因为: %w", err)
	}

	return result, nil
}

var _ tool.Tool = (*ListTextFilesTool)(nil)
