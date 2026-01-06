package database

import (
	"database/sql"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"skeleton-go/pkg/logger"
)

var gormOpen = gorm.Open

// NewGORM creates a GORM instance backed by the provided *sql.DB connection.
func NewGORM(db *sql.DB, log *logger.Logger, logQueries bool) (*gorm.DB, error) {
	dialector := mysql.New(mysql.Config{Conn: db})

	gormLogger := newGormLogger(log, logQueries)

	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	return gormOpen(dialector, gormConfig)
}
