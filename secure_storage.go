package secure_storage

import (
	"github.com/jinzhu/gorm"
)

type Secured interface {
	EnableSecureStorage()
}

type Enable struct{}

func (h Enable) EnableSecureStorage() {}

func Init(db *gorm.DB, c Config) error {
	config = c

	db.Callback().Query().Before("gorm:query").Register("secure_storage:query", beforeQuery)
	db.Callback().Query().After("gorm:query").Register("secure_storage:after_query", afterQuery)
	db.Callback().Create().Before("gorm:before_create").Register("secure_storage:create", beforeCreate)
	db.Callback().Update().Before("gorm:assign_updating_attributes").Register("secure_storage:update", beforeUpdate)
	db.Callback().Update().After("gorm:after_update").Register("secure_storage:after_update", afterUpdate)
	return nil
}

func HaveNotSecureStorage(scope *gorm.Scope) bool {
	_, ok := scope.Value.(Secured)
	return !ok || len(getFields(scope)) == 0
}
