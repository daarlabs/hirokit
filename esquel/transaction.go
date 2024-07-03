package esquel

import (
	"context"
	"time"
)

func (d *DB) Begin() (*DB, error) {
	ctx, cancel := context.WithCancel(d.context)
	db := &DB{
		context:     ctx,
		cancel:      cancel,
		DB:          d.DB,
		driverName:  d.driverName,
		transaction: true,
		rollback:    false,
		log:         d.log,
		timeout:     d.timeout,
	}
	q := "BEGIN;"
	t := time.Now()
	_, err := d.DB.ExecContext(ctx, q)
	log(db.log, q, time.Now().Sub(t))
	return db, err
}

func (d *DB) Rollback() error {
	if !d.transaction {
		return nil
	}
	if d.cancel != nil {
		defer d.cancel()
	}
	d.rollback = true
	q := "ROLLBACK;"
	t := time.Now()
	_, err := d.DB.ExecContext(d.context, q)
	log(d.log, q, time.Now().Sub(t))
	d.transaction = false
	return err
}

func (d *DB) Commit() error {
	if !d.transaction {
		return nil
	}
	if d.cancel != nil {
		defer d.cancel()
	}
	q := "COMMIT;"
	t := time.Now()
	_, err := d.DB.ExecContext(d.context, q)
	log(d.log, q, time.Now().Sub(t))
	d.transaction = false
	return err
}

func (d *DB) MustBegin() *DB {
	db, err := d.Begin()
	if err != nil {
		panic(err)
	}
	d.transaction = true
	return db
}

func (d *DB) MustRollback() {
	if !d.transaction {
		return
	}
	err := d.Rollback()
	if err != nil {
		panic(err)
	}
}

func (d *DB) MustCommit() {
	if !d.transaction {
		return
	}
	if err := d.Commit(); err != nil {
		panic(err)
	}
}
