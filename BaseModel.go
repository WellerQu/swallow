package swallow

import (
	"reflect"
	"time"
)

// BaseModel 基础模型
type BaseModel struct {
	InsertAt    int64 `swallow:"column:insert_at;" json:"insertAt"`
	UpdateAt    int64 `swallow:"column:update_at;" json:"updateAt"`
	DeleteAt    int64 `swallow:"column:delete_at;" json:"deleteAt"`
	IsDeleted   bool  `swallow:"column:is_deleted;" json:"-"`
	IsNewRecord bool  `swallow:"-;" json:"-"`
}

// TableName 实现 desc 接口, 返回表名
func (m *BaseModel) TableName() string {
	return reflect.TypeOf(m).Elem().Name()
}

// 实现 hooks 接口, 在查询前做点什么
func (m *BaseModel) beforeFind() {
	m.IsDeleted = false
}

// 实现 hooks 接口, 在修改前, 添加修改时间, 实际上数据库也能做
func (m *BaseModel) beforeSave() {
	m.UpdateAt = time.Now().UnixNano() / int64(time.Millisecond)
}

// 实现 hooks 接口, 在插入前, 添加插入时间, 实际上数据库也能做
func (m *BaseModel) beforeCreate() {
	m.InsertAt = time.Now().UnixNano() / int64(time.Millisecond)
}

// 实现 hooks 接口, 在删除前, 添加删除时间, 实际上数据库也能做
func (m *BaseModel) beforeDelete() {
	// m.DeleteAt = time.Now().UnixNano() / int64(time.Millisecond)
}
