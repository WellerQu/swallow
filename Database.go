package swallow

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	// 导入mysql数据库驱动
	_ "github.com/go-sql-driver/mysql"
)

// DB 数据库操作
type DB interface {
	Begin() error
	Commit() error
	Rollback() error
	Find(model interface{}) ([]interface{}, error)
	First(model interface{}) (interface{}, error)
	Create(model interface{}) (int64, error)
	Save(model interface{}) (int64, error)
	Delete(model interface{}) (int64, error)
	Close() error
}

type database struct {
	conn *sql.DB
	tx   *sql.Tx
}

// Open 打开一个数据库连接, 并返回一个可用于操作数据库的connection对象
func Open(connectionString string) (DB, error) {
	db := new(database)
	conn, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	db.conn = conn
	return db, nil
}

// Close 关闭一个数据库连接
func Close(db DB) error {
	return db.Close()
}

// Close 关闭数据库连接
func (db *database) Close() error {
	return db.conn.Close()
}

// Begin 打开事务
func (db *database) Begin() error {
	if db.tx != nil {
		return errors.New("已经开启过事务")
	}

	tx, err := db.conn.Begin()
	db.tx = tx

	return err
}

// Commit 提交事务
func (db *database) Commit() error {
	if db.tx == nil {
		return errors.New("还没有开启过事务")
	}

	err := db.tx.Commit()
	db.tx = nil

	return err
}

// Rollback 回滚事务
func (db *database) Rollback() error {
	if db.tx == nil {
		return errors.New("还没有开启过事务")
	}

	err := db.tx.Rollback()
	db.tx = nil

	return err
}

// Find 查询列表记录
func (db *database) Find(model interface{}) (list []interface{}, err error) {
	if cv, ok := model.(hooks); ok {
		cv.beforeFind()
	}

	fields := getFieldsByModel(model)
	tableName := getTableNameByModel(model)

	query := new(querier)

	SQLBuilder := new(strings.Builder)
	SQLBuilder.WriteString("select ")

	columns := []string{}

	for _, f := range fields {
		if f.isIgnore() {
			continue
		}

		columns = append(columns, fmt.Sprintf("`%s`", f.columnName()))

		if f.isBlank() {
			continue
		}

		query.Where(fmt.Sprintf("`%s` = ?", f.columnName()), f.reflectValue.Interface())
	}

	SQLBuilder.WriteString(strings.Join(columns, ", "))
	SQLBuilder.WriteString(" from ")
	SQLBuilder.WriteString(tableName)

	query.Prepare(SQLBuilder.String())

	rows, err := query.Query(db)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		fieldMap := make(map[string]Field)
		record := reflect.New(reflect.TypeOf(model).Elem()).Interface()
		fields := getFieldsByModel(record)

		for _, f := range fields {
			fieldMap[f.columnName()] = f
		}

		scanArgs := make([]interface{}, len(columns))
		values := make([][]uint8, len(columns))

		for i := range values {
			scanArgs[i] = &values[i]
		}

		scanErr := rows.Scan(scanArgs...)
		if scanErr != nil {
			break
		}

		for i, c := range columns {
			field := fieldMap[strings.Trim(c, "`")]
			buffer := values[i]

			if !field.reflectValue.CanSet() {
				continue
			}

			switch field.reflectValue.Kind() {
			case reflect.String:
				field.reflectValue.SetString(string(buffer))
			case reflect.Int64, reflect.Int:
				byteToInt, _ := strconv.Atoi(string(buffer))
				field.reflectValue.SetInt(int64(byteToInt))
			case reflect.Bool:
				field.reflectValue.SetBool(buffer[0] == 0)
			}
		}

		list = append(list, record)
	}

	return
}

// Create 创建新纪录
func (db *database) Create(model interface{}) (n int64, execErr error) {
	if cv, ok := model.(hooks); ok {
		cv.beforeCreate()
	}

	fields := getFieldsByModel(model)
	tableName := getTableNameByModel(model)

	SQLBuilder := new(strings.Builder)
	SQLBuilder.WriteString("insert into ")
	SQLBuilder.WriteString(tableName)

	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}

	for _, v := range fields {
		if v.isBlank() {
			continue
		}
		if v.isIgnore() {
			continue
		}

		columns = append(columns, fmt.Sprintf("`%s`", v.columnName()))
		placeholders = append(placeholders, "?")
		values = append(values, v.reflectValue.Interface())
	}

	SQLBuilder.WriteString(" (")
	SQLBuilder.WriteString(strings.Join(columns, ", "))
	SQLBuilder.WriteString(")")
	SQLBuilder.WriteString(" values (")
	SQLBuilder.WriteString(strings.Join(placeholders, ", "))
	SQLBuilder.WriteString(")")

	exec := new(executor)
	result, execErr := exec.Prepare(SQLBuilder.String(), values...).Exec(db)

	if execErr != nil {
		return 0, execErr
	}

	lastInsertID, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	if primaryKeyField, ok := getPrimaryKeyByModel(model); ok {
		primaryKeyField.reflectValue.SetInt(lastInsertID)
	}

	return result.RowsAffected()
}

// First 查询第一条符合条件的录
func (db *database) First(model interface{}) (one interface{}, err error) {
	list, findErr := db.Find(model)
	if len(list) == 0 {
		return nil, findErr
	}

	return list[0], nil
}

// Save 保存更新
func (db *database) Save(model interface{}) (n int64, err error) {
	if cv, ok := model.(hooks); ok {
		cv.beforeSave()
	}

	fields := getFieldsByModel(model)
	tableName := getTableNameByModel(model)

	exec := new(executor)

	SQLBuilder := new(strings.Builder)
	SQLBuilder.WriteString("update ")
	SQLBuilder.WriteString(tableName)
	SQLBuilder.WriteString(" set ")

	values := []interface{}{}
	columns := []string{}

	for _, f := range fields {
		if f.isBlank() {
			continue
		}
		if f.isIgnore() {
			continue
		}

		if f.isPrimaryKey() {
			exec.Where(fmt.Sprintf("`%s` = ?", f.columnName()), f.reflectValue.Interface())
		} else {
			columns = append(columns, fmt.Sprintf("`%s` = ?", f.columnName()))
			values = append(values, f.reflectValue.Interface())
		}
	}

	SQLBuilder.WriteString(strings.Join(columns, ", "))

	result, err := exec.Prepare(SQLBuilder.String(), values...).Exec(db)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete 删除记录
func (db *database) Delete(model interface{}) (n int64, err error) {
	if cv, ok := model.(hooks); ok {
		cv.beforeDelete()
	}

	exec := new(executor)

	fields := getFieldsByModel(model)
	tableName := getTableNameByModel(model)

	SQLBuilder := new(strings.Builder)
	SQLBuilder.WriteString("delete from ")
	SQLBuilder.WriteString(tableName)

	for _, f := range fields {
		if f.isBlank() {
			continue
		}
		if f.isIgnore() {
			continue
		}

		exec.Where(fmt.Sprintf("`%s` = ?", f.columnName()), f.reflectValue.Interface())
	}

	result, err := exec.Prepare(SQLBuilder.String()).Exec(db)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
