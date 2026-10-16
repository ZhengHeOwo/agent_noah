package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository"

	_ "modernc.org/sqlite"
)

// AddSerial 为context_tables更新序号.
func (s *Sqlite) AddSerial() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("AddSerial 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if err := s.UpdateRuntimeConfig(tx, "serial", s.Serial+1); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("AddSerial出现事务提交异常 %w", err)
	}

	s.Serial++

	return nil
}

// GetNewProfile 更新并获取context_tables新的profile值.
func (s *Sqlite) GetNewProfile() (int64, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return -1, fmt.Errorf("GetNewProfile 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	now := s.Profile
	new := now + 1
	if new >= 3 {
		new = 1
	}
	if err := s.UpdateRuntimeConfig(tx, "context_profile", new); err != nil {
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, fmt.Errorf("GetNewProfile出现事务提交异常 %w", err)
	}

	s.Profile = new

	return new, nil
}

// InsertMessage 根据当前序号/ID将信息插入数据库.
func (s *Sqlite) InsertMessage(message *model.Message, nowSteps int64) error {
	if s.Serial < 1 {
		return fmt.Errorf("当前序列为异常值 %d", s.Serial)
	}

	timeText := time.Now().Format(time.RFC3339)

	message.Prefix = model.Prefix{
		Serial:         s.Serial,
		Id:             s.Id,
		OccurrenceTime: timeText,
		MaxModelSteps:  repository.MaxModelSteps,
		NowSteps:       nowSteps,
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("message编码失败 %w", err)
	}

	jsonText := string(jsonBytes)

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("InsertMessage 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"INSERT INTO active_context (profile, serial, id, message) VALUES (?, ?, ?, ?)",
		s.Profile,
		s.Serial,
		s.Id,
		jsonText,
	); err != nil {
		return fmt.Errorf("数据插入active_context失败 %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO history_context (profile, serial, id, message) VALUES (?, ?, ?, ?)",
		s.Profile,
		s.Serial,
		s.Id,
		jsonText,
	); err != nil {
		return fmt.Errorf("数据插入history_context失败 %w", err)
	}

	s.Id++
	if err := s.UpdateRuntimeConfig(tx, "id", s.Id); err != nil {
		log.Println(err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("InsertMessage出现事务提交异常 %w", err)
	}

	return nil
}

func (s *Sqlite) AddModelSteps(r *repository.RuntimeConfig) {
	r.NowSteps++
	return
}

const limit = 10

// LoadRecentContext 重新进入活跃态时获取一定数量的活跃上下文, 并将运行器状态更新为active.
func (s *Sqlite) LoadRecentContext(contextRecord *repository.ContextRecord, system string) error {
	nowProfile := s.Profile
	newProfile, err := s.GetNewProfile()
	if err != nil {
		return err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("LoadRecentContext 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
	DELETE FROM active_context
	WHERE profile = ?;
	`, newProfile); err != nil {
		return fmt.Errorf("LoadRecentContext [active_context new_profile: %d]清理失败 %w", nowProfile, err)
	}

	if _, err := tx.Exec(`
	INSERT INTO active_context (profile, serial, id, message)
	SELECT ?, serial, id, message
	FROM active_context
	WHERE profile = ?
	AND serial IN (
	SELECT DISTINCT serial
	FROM active_context
	WHERE profile = ?
	ORDER BY serial DESC
	LIMIT ?
	)`, newProfile, nowProfile, nowProfile, limit); err != nil {
		return fmt.Errorf("LoadRecentContext [active_context profile:%d -> profile:%d]转存失败 %w", nowProfile, newProfile, err)
	}

	if _, err := tx.Exec(`
	DELETE FROM active_context
	WHERE profile = ?;
	`, nowProfile); err != nil {
		return fmt.Errorf("LoadRecentContext [active_context old_profile: %d]清理失败 %w", nowProfile, err)
	}

	if _, err := tx.Exec(`
	DELETE FROM history_context
	WHERE profile = ?;
	`, nowProfile); err != nil {
		return fmt.Errorf("LoadRecentContext [history_context old_profile: %d]清理失败 %w", nowProfile, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("LoadRecentContext 事务提交异常 %w", err)
	}

	return s.GetActiveContext(contextRecord, system)
}

// GetActiveContext 获取当前profile的全部活跃上下文.
func (s *Sqlite) GetActiveContext(contextRecord *repository.ContextRecord, system string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("GetActiveContext 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
	SELECT message
	FROM active_context
	WHERE profile = ?
	ORDER BY id ASC
	`, s.Profile)
	if err != nil {
		return fmt.Errorf("GetActiveContext查询异常 %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var payload string
		var record model.Message
		if err := rows.Scan(&payload); err != nil {
			return fmt.Errorf("LoadRecentContext取出当前行异常 %w", err)
		}
		if err := json.Unmarshal([]byte(payload), &record); err != nil {
			return fmt.Errorf("LoadRecentContext解码message异常 %w", err)
		}
		contextRecord.Messages = append(contextRecord.Messages, record)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("LoadRecentContext循环结束异常 %w", err)
	}

	if err := s.ComplementMessages(contextRecord, system); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("GetActiveContext 事务提交异常 %w", err)
	}

	return nil
}

// ComplementMessages 负责插入提示词并做触发检查.
func (s *Sqlite) ComplementMessages(contextRecord *repository.ContextRecord, system string) error {
	if contextRecord.Messages[0].Role == model.RoleSystem {
		return fmt.Errorf("Runtime上下文逻辑异常, 执行插入前messages[0]的Role不应为RoleSystem")
	}

	contextRecord.Messages = slices.Insert(contextRecord.Messages, 0, model.Message{
		Role:    model.RoleSystem,
		Content: strings.TrimSpace(system),
	})

	checkPointRole := contextRecord.Messages[len(contextRecord.Messages)-1].Role
	if checkPointRole != model.RoleUser && checkPointRole != model.RoleTool {
		return fmt.Errorf("Runtime上下文逻辑异常, 本次运行的触发者非预期")
	}

	return nil
}

// DeleteActiveContextBySerialAndID 根据提供的序号和id删除当前活跃上下文的一个指定message.
func (s *Sqlite) DeleteActiveContextBySerialAndID(serial, id int) (string, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("删除事务创建失败 %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		DELETE FROM active_context
		WHERE profile = ? AND serial = ? AND id = ?
	`, s.Profile, serial, id)
	if err != nil {
		return "", fmt.Errorf("删除 serial %d | id %d 失败: %w", serial, id, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("获取删除数量失败: %w", err)
	}

	if count != 1 {
		return "", fmt.Errorf("删除 serial %d, id %d 存在程序异常, 影响数量不为1: %w", serial, id, err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("删除事务提交失败: %w", err)
	}

	return fmt.Sprintf("删除 serial %d | id %d 成功", serial, id), nil
}

// GetRuntimeConfig 获取当前Runtime的运行配置.
func (s *Sqlite) GetRuntimeConfig(runtimeProfile int64) (*repository.RuntimeConfig, error) {
	runtimeConfig := &repository.RuntimeConfig{}

	err := s.DB.QueryRow(`
	SELECT context_profile, status, serial, id
	FROM runtime_config 
	WHERE profile = ?
	`, runtimeProfile).Scan(
		&runtimeConfig.ContextProfile,
		&runtimeConfig.Status,
		&runtimeConfig.Serial,
		&runtimeConfig.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("运行配置读取失败: %w", err)
	}

	runtimeConfig.NowSteps = 1

	return runtimeConfig, nil
}

// UpdateRuntimeConfig 更新Runtime的运行配置.
func (s *Sqlite) UpdateRuntimeConfig(tx *sql.Tx, column string, argument any) error {
	column = strings.TrimSpace(column)
	var query string

	switch column {
	case "status":
		query = "UPDATE runtime_config SET status = ?"
	case "serial":
		query = "UPDATE runtime_config SET serial = ?"
	case "id":
		query = "UPDATE runtime_config SET id = ?"
	case "context_profile":
		query = "UPDATE runtime_config SET context_profile = ?"
	case "":
		return fmt.Errorf("column不得为空")
	default:
		return fmt.Errorf("不支持更新列 %q", column)
	}

	if _, err := tx.Exec(query, argument); err != nil {
		return fmt.Errorf("runtime_config更新失败 %w", err)
	}

	return nil
}

func (s *Sqlite) StatusTransition() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("StatusTransition 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if err := s.UpdateRuntimeConfig(tx, "status", "active"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("StatusTransition 出现事务提交异常 %w", err)
	}

	return nil
}

func (s *Sqlite) SqlDB() *sql.DB