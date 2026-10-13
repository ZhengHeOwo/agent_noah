package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// NewDB 使用sqlite获取句柄.
func NewDB(databasePath string) (*sql.DB, error) {
	databasePath = strings.TrimSpace(databasePath)
	if databasePath == "" {
		return nil, fmt.Errorf("databasePath 不可为空")
	}

	dir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create database directory %q: %w", databasePath, err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("数据库入口创建失败 %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库Ping失败 %w", err)
	}

	if err := createAllTables(db); err != nil {
		return nil, err
	}

	if err := initAllTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

// initTables sqlite事务建表集合.
func createAllTables(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("createAllTables 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS active_context (
    profile INTEGER NOT NULL CHECK (profile IN (1, 2)),
    serial INTEGER NOT NULL,
    id INTEGER NOT NULL,
    message TEXT NOT NULL,
    PRIMARY KEY (profile, serial, id)
) WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("active_context TABLE创建失败 %w", err)
	}

	if _, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS history_context (
    profile INTEGER NOT NULL CHECK (profile IN (1, 2)),
    serial INTEGER NOT NULL,
    id INTEGER NOT NULL,
    message TEXT NOT NULL,
    PRIMARY KEY (profile, serial, id)
) WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("history_context TABLE创建失败 %w", err)
	}

	if _, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS runtime_config (
	profile  INTEGER NOT NULL PRIMARY KEY,
	context_profile INTEGER NOT NULL CHECK (context_profile IN (1, 2)),
	status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
	serial INTEGER NOT NULL,
	id INTEGER NOT NULL
	)WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("runtime_config TABLE创建失败 %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createAllTables 事务提交异常 %w", err)
	}

	return nil
}

// initAllTables 根据后续逻辑需要进行初始化.
func initAllTables(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("initAllTables 事务创建失败 %w", err)
	}
	defer tx.Rollback()

	if err := checkDatabaseStatus(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("initAllTables 事务提交异常 %w", err)
	}

	return nil
}

const (
	defaultRuntimeProfile int64  = 1
	defaultContextProfile int64  = 1
	defaultRuntimeStatus  string = "inactive"
	defaultSerial         int64  = 1
	defaultID             int64  = 1
)

// checkDatabaseStatus 初始化运行所需配置.
func checkDatabaseStatus(tx *sql.Tx) error {
	var profileContext, serial, id int64
	var status string

	err := tx.QueryRow(`
	SELECT context_profile, status, serial, id
	FROM runtime_config
	WHERE profile = ?
`, defaultRuntimeProfile).Scan(&profileContext, &status, &serial, &id)

	if errors.Is(err, sql.ErrNoRows) {
		if err := initRuntimeConfiguration(tx); err != nil {
			return err
		}

	} else if err != nil {
		return fmt.Errorf("读取运行配置失败: %w", err)
	} else {
		var errs []error

		if serial < 1 {
			errs = append(errs, fmt.Errorf("[当前 profile %d]下的serial为异常值: %d", 1, serial))
		}

		if id < 1 {
			errs = append(errs, fmt.Errorf("[当前 profile %d]下的id为异常值: %d", 1, id))
		}

		if status != "active" && status != "inactive" {
			errs = append(errs, fmt.Errorf("[当前 profile %s]下的status为异常值: %s", 1, status))
		}

		if len(errs) != 0 {
			log.Printf("\n当前配置错误: %v\n开始尝试切库\n", errors.Join(errs...))

			//未写
		}
	}

	return nil
}

func initRuntimeConfiguration(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO runtime_config (profile, context_profile, status, serial, id) VALUES (?, ?, ?, ?, ?)
	`,
		defaultRuntimeProfile,
		defaultContextProfile,
		defaultRuntimeStatus,
		defaultSerial,
		defaultID,
	); err != nil {
		return fmt.Errorf("初始化 runtime_config 失败: %w", err)
	}
	return nil
}
