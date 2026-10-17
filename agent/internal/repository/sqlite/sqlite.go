package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/ZhengHeOwo/agent_noah/agent/internal/repository"
)

type Sqlite struct {
	db *sql.DB
}

// 提供唯一创建函数.
func BuildSqlite() (*Sqlite, error) {

}