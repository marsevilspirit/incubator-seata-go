package db

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/seata/seata-go/pkg/util/log"
)

const TimeLayout = "2006-01-02 15:04:05.999999999-07:00"

type ExecStatement[T any] func(obj T, stmt *sql.Stmt) (int64, error)

type ScanRows[T any] func(rows *sql.Rows) (T, error)

type Store struct {
	db *sql.DB
}

func SelectOne[T any](db *sql.DB, sql string, fn ScanRows[T], args ...any) (T, error) {
	var result T
	log.Debugf("Preparing SQL: %s", sql)
	stmt, err := db.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return result, err
	}

	log.Debugf("setting params to Stmt: %v", args)
	rows, err := stmt.Query(args...)
	defer rows.Close()
	if err != nil {
		return result, nil
	}

	if rows.Next() {
		return fn(rows)
	}
	return result, errors.New("no target selected")
}

func SelectList[T any](db *sql.DB, sql string, fn ScanRows[T], args ...any) ([]T, error) {
	result := make([]T, 0)

	log.Debugf("Preparing SQL: %s", sql)
	stmt, err := db.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return result, err
	}

	log.Debugf("setting params to Stmt: %v", args)
	rows, err := stmt.Query(args...)
	defer rows.Close()
	if err != nil {
		return result, err
	}

	for rows.Next() {
		obj, err := fn(rows)
		if err != nil {
			return result, err
		}
		result = append(result, obj)
	}

	return result, nil
}

func ExecuteUpdate[T any](db *sql.DB, sql string, fn ExecStatement[T], obj T) (int64, error) {
	log.Debugf("Preparing SQL: %s", sql)
	stmt, err := db.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return 0, err
	}

	log.Debugf("setting params to Stmt: %v", obj)

	rowsAffected, err := fn(obj, stmt)
	if err != nil {
		return rowsAffected, err
	}

	return rowsAffected, nil
}

func ExecuteUpdateArgs(db *sql.DB, sql string, args ...any) (int64, error) {
	log.Debugf("Preparing SQL: %s", sql)
	stmt, err := db.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return 0, err
	}

	log.Debugf("setting params to Stmt: %v", args)

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}