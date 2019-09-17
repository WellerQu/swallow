package swallow

import (
	"database/sql"
	"log"
	"strings"
)

type executor struct {
	statement string
	condition []string
	args0     []interface{}
	args1     []interface{}
}

func (e *executor) Prepare(statement string, args ...interface{}) *executor {
	if e.args0 == nil {
		e.args0 = []interface{}{}
	}

	e.statement = statement
	e.args0 = append(e.args0, args...)
	return e
}

func (e *executor) Where(condition string, args ...interface{}) *executor {
	if e.args1 == nil {
		e.args1 = []interface{}{}
	}

	e.condition = append(e.condition, condition)
	e.args1 = append(e.args1, args...)
	return e
}

func (e *executor) getSQL() string {
	builder := new(strings.Builder)
	builder.WriteString(e.statement)

	// Q:这里其实一个strings.Join就搞定的问题, 为什么要写成这样?
	// A:为了性能, 因此采用strings.Builder, 参见 strings/builder.go源码
	if length := len(e.condition); length > 0 {
		builder.WriteString(" where ")

		for i, v := range e.condition {
			builder.WriteString(v)
			if i != length-1 {
				builder.WriteString(" and ")
			}
		}
	}

	// 至此, 拼接出了一个类似于这样的SQL
	// select id, name from student where id = ?
	// 或者
	// update student set name = ? where id = ?
	return builder.String()
}

func (e *executor) getSQLArgs() []interface{} {
	// 汇聚所有的参数
	args := []interface{}{}
	args = append(args, e.args0...)
	args = append(args, e.args1...)

	return args
}

func (e *executor) Exec(db *database) (sql.Result, error) {
	sql := e.getSQL()
	args := e.getSQLArgs()

	return e.ExecRaw(db, sql, args...)
}

func (e *executor) ExecRaw(db *database, sqlStatement string, args ...interface{}) (sql.Result, error) {
	log.Println("[SWALLOW]", sqlStatement, args)

	var stmt *sql.Stmt
	var err error

	if db.tx != nil {
		stmt, err = db.tx.Prepare(sqlStatement)
	} else {
		stmt, err = db.conn.Prepare(sqlStatement)
	}

	if err != nil {
		return nil, err
	}

	return stmt.Exec(args...)
}
