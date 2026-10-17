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


func (s *Sqlite) withTx(ctx context.Context, fn func(tx *sql.Tx) error ) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 
	}
}