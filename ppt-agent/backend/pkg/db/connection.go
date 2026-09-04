package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cloudwego/ppt-agent/pkg/db/internal/schema"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

// DB is the shared GORM handle initialized during service startup.
var DB *gorm.DB

const dbOperationTimeout = 5 * time.Second

func withOperationTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbOperationTimeout)
}

// Init 使用生产级设置初始化数据库连接。
// - maxOpenConns/maxIdleConns：限制并发数据库负载
// - connMaxLifetime：在 MySQL 的 wait_timeout 过期前回收连接
// - connMaxIdleTime：关闭空闲时间超过 MySQL 的 interactive_timeout 的连接
// - tcpKeepalive：在 TCP 层检测死连接，而无需等待 MySQL 超时
func Init(dsn string) error {
	dsn, err := dsnWithTimeouts(dsn)
	if err != nil {
		return fmt.Errorf("parse MySQL DSN: %w", err)
	}

	DB, err = gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("DB ping failed: %w", err)
	}
	var currentDatabase string
	if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&currentDatabase); err != nil {
		return fmt.Errorf("resolve current database: %w", err)
	}
	if !isBusinessDatabase(currentDatabase) {
		return fmt.Errorf("refusing to migrate non-business database %q", currentDatabase)
	}

	if err := schema.Migrate(DB); err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}

	logger.Info("mysql_connected")
	return nil
}

func isBusinessDatabase(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "information_schema", "mysql", "performance_schema", "sys":
		return false
	default:
		return true
	}
}

func dsnWithTimeouts(dsn string) (string, error) {
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	cfg.InterpolateParams = true
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	return cfg.FormatDSN(), nil
}
