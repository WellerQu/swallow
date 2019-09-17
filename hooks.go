package swallow

type hooks interface {
	beforeFind()
	beforeSave()
	beforeCreate()
	beforeDelete()
}
