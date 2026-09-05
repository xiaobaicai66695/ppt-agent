package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// InstanceLock is a MySQL advisory lock held by one web process. The current
// runtime uses local task state, SSE listeners and local output files, so
// active-active processes would split ownership and corrupt delivery.
type InstanceLock struct {
	conn *sql.Conn
	name string
	once sync.Once
}

func AcquireInstanceLock(ctx context.Context, name string) (*InstanceLock, error) {
	if DB == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var acquired sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 0)", name).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		_ = conn.Close()
		return nil, fmt.Errorf("another ppt-agent instance already owns lock %q", name)
	}
	return &InstanceLock{conn: conn, name: name}, nil
}

func (l *InstanceLock) Release(ctx context.Context) error {
	if l == nil || l.conn == nil {
		return nil
	}
	var releaseErr error
	l.once.Do(func() {
		var released sql.NullInt64
		releaseErr = l.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", l.name).Scan(&released)
		if err := l.conn.Close(); releaseErr == nil {
			releaseErr = err
		}
	})
	return releaseErr
}
