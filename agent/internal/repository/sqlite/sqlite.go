package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository"
)

type Sqlite struct {
	DB      *sql.DB
	Profile int64
	Serial  int64
	Id      int64
}

func NewSqlite(db *sql.DB, pendinContextRecord *repository.ContextRecord) (*Sqlite, error) {
	if db == nil {
		return nil, fmt.Errorf("sqlite db不可为空")
	}

	sqlite := &Sqlite{
		DB: db,
	}

	runtimeConfig, err := sqlite.GetRuntimeConfig(1)
	if err != nil {
		return nil, err
	}

	sqlite.Profile = runtimeConfig.ContextProfile
	sqlite.Serial = runtimeConfig.Serial
	sqlite.Id = runtimeConfig.Id

	if pendinContextRecord != nil {
		// 一会放将Message历史转载到当前库逻辑
	}

	return sqlite, nil
}

func (s *Sqlite) Close() error {
	return s.DB.Close()
}
