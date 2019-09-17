package swallow

import (
	"database/sql"
	"log"
)

type querier struct {
	executor
}

func (q *querier) Query(db *database) (*sql.Rows, error) {
	sql := q.getSQL() + " limit 20"
	args := q.getSQLArgs()

	return q.QueryRaw(db, sql, args...)
}

func (q *querier) QueryRaw(db *database, sql string, args ...interface{}) (*sql.Rows, error) {
	log.Println("[SWALLOW]", sql, args)

	if db.tx != nil {
		return db.tx.Query(sql, args...)
	}

	return db.conn.Query(sql, args...)
}
