package esquel

import (
	"context"
	"database/sql"
	"time"
)

type DB struct {
	*sql.DB
	context     context.Context
	cancel      context.CancelFunc
	driverName  string
	transaction bool
	rollback    bool
	log         bool
	timeout     time.Duration
	logFunc     func(time string, query string, args []any)
}

const (
	Postgres = "postgres"
	Mysql    = "mysql"
)

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return wrapConnection(db, driverName), err
}

func wrapConnection(db *sql.DB, driverName string) *DB {
	return &DB{
		context:     context.Background(),
		DB:          db,
		driverName:  driverName,
		transaction: false,
		rollback:    false,
		log:         false,
	}
}

func (d *DB) Q(query string, arg ...Map) *Esquel {
	return New(d).Q(query, arg...)
}

func (d *DB) DriverName() string {
	return d.driverName
}

func (d *DB) Log(logFunc func(time string, query string, args []any)) *DB {
	d.logFunc = logFunc
	return d
}
