package repository

import (
	"database/sql"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/model"
)

const MaxModelSteps = 30

type ContextRecord struct {
	Messages []model.Message
}

// RuntimeConfig 运行时配置.
type RuntimeConfig struct {
	ContextProfile int64
	MaxModelSteps  int64
	NowSteps       int64
	Status         string
	Serial         int64
	Id             int64
}

type ContextControl interface {
	AddSerial() error
	GetNewProfile() (int64, error)
	InsertMessage(message *model.Message, nowSteps int64) error
	AddModelSteps(r *RuntimeConfig)
	LoadRecentContext(contextRecord *ContextRecord, system string) error
	GetActiveContext(contextRecord *ContextRecord, system string) error
	ComplementMessages(contextRecord *ContextRecord, system string) error
	DeleteActiveContextBySerialAndID(serial, id int) (string, error)
	GetRuntimeConfig(runtimeProfile int64) (*RuntimeConfig, error)
	UpdateRuntimeConfig(tx *sql.Tx, column string, argument any) error
	StatusTransition() error
}
