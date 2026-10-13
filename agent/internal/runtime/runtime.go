package runtime

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/tool"
)

// Runtime 管理对话历史和单轮内的多步工具调用。
type Runtime struct {
	llm            model.Model
	contextControl repository.ContextControl
	runtimeConfig  *repository.RuntimeConfig
	modelName      string
	systemPrompt   string
	tools          *tool.ToolRegistry
}

// NewRuntime 创建Agent运行时。
func NewRuntime(llm model.Model, contextControl repository.ContextControl, modelName string, systemPrompt string, tools *tool.ToolRegistry) (*Runtime, error) {
	if llm == nil {
		return nil, fmt.Errorf("Model must not be empty")
	}

	if strings.TrimSpace(modelName) == "" {
		return nil, fmt.Errorf("ModelName must not be empty")
	}

	if strings.TrimSpace(systemPrompt) == "" {
		return nil, fmt.Errorf("systemPrompt must not be empty")
	}

	if tools == nil {
		return nil, fmt.Errorf("tool registry must not be empty")
	}

	runtimeConfig, err := contextControl.GetRuntimeConfig(1)
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		llm:            llm,
		contextControl: contextControl,
		runtimeConfig:  runtimeConfig,
		modelName:      strings.TrimSpace(modelName),
		systemPrompt:   systemPrompt,
		tools:          tools,
	}

	return runtime, nil
}

// RunTurn 执行一轮运行器。
func (r *Runtime) RunTurn(ctx context.Context, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("input 不能为空")
	}

	if err := r.contextControl.InsertMessage(&model.Message{
		Role:    model.RoleUser,
		Content: input,
	}, r.runtimeConfig.NowSteps); err != nil {
		return "", err
	}

	for step := 1; step <= repository.MaxModelSteps; step++ {
		contextRecord := &repository.ContextRecord{}

		if r.runtimeConfig.Status == "inactive" {
			if err := r.contextControl.LoadRecentContext(contextRecord, r.systemPrompt); err != nil {
				return "", err
			}
		} else {
			if err := r.contextControl.GetActiveContext(contextRecord, r.systemPrompt); err != nil {
				return "", err
			}
		}

		response, err := r.llm.Complete(ctx, model.Request{
			Model:    r.modelName,
			Messages: contextRecord.Messages,
			Tools:    r.tools.Definitions(),
		})
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrModelInvocationFailed, err)
		}

		message := response.Message
		if message.Role != model.RoleAssistant {
			return "", fmt.Errorf("%w: want %q, got: %q", ErrResponseRoleError, model.RoleAssistant, message.Role)
		}

		if len(message.ToolCalls) == 0 {
			if strings.TrimSpace(message.Content) == "" {
				return "", fmt.Errorf("%w: finish_reason=%q", ErrEmptyContent, response.FinishReason)
			}

			if err := r.contextControl.InsertMessage(&message, r.runtimeConfig.NowSteps); err != nil {
				return "", err
			}

			if err := r.contextControl.StatusTransition(); err != nil {
				return "", err
			}

			r.contextControl.AddSerial()
			return message.Content, nil
		}

		if err := r.contextControl.InsertMessage(&message, r.runtimeConfig.NowSteps); err != nil {
			return "", err
		}

		for _, call := range message.ToolCalls {
			log.Printf("调用工具 %s", call.Name)

			if strings.TrimSpace(call.Name) == "" {
				return "", fmt.Errorf("%w: 工具名为空", ErrInvalidToolCall)
			}

			if strings.TrimSpace(call.ID) == "" {
				return "", fmt.Errorf("%w: 工具调用ID为空", ErrInvalidToolCall)
			}

			toolCtx, cancel := context.WithTimeout(ctx, 300*time.Second)
			defer cancel()

			result := r.executeToolCall(toolCtx, call)

			log.Printf("工具 %s 调用结束", call.Name)

			toolResultMessage := &model.Message{
				Role:       model.RoleTool,
				Content:    result,
				ToolCallID: call.ID,
			}

			if err := r.contextControl.InsertMessage(toolResultMessage, r.runtimeConfig.NowSteps); err != nil {
				return "", err
			}
		}

		r.contextControl.AddModelSteps(r.runtimeConfig)
	}

	return "", ErrMaxStepsExceeded
}

func (r *Runtime) executeToolCall(toolCtx context.Context, call model.ToolCall) string {
	registeredTool, exists := r.tools.Get(call.Name)
	if !exists {
		return fmt.Sprintf("工具执行失败, 因为: 未注册%q", call.Name)
	}

	result, err := registeredTool.Execute(toolCtx, call.Arguments)
	if err != nil {
		if result == "" {
			return err.Error()
		}
		return fmt.Sprintf("%v\n已返回的部分结果:\n%s", err, result)
	}
	return result
}
