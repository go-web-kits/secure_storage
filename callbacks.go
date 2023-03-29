package secure_storage

import (
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
)

func beforeQuery(scope *gorm.Scope) {
	// pp.Println(scope)
}

func afterQuery(scope *gorm.Scope) {
	if guard(scope) {
		return
	}
	CheckDecryptAndFill(scope)
}

func beforeCreate(scope *gorm.Scope) {
	if guard(scope) {
		return
	}
	EncryptAndFill(scope)
}

func beforeUpdate(scope *gorm.Scope) {
	if guard(scope) || HaveNotSecureStorage(scope) {
		return
	}
	if attrs, ok := scope.InstanceGet("gorm:update_interface"); ok {
		reflectAttrs := reflect.ValueOf(attrs)
		if reflectAttrs.Kind() != reflect.Map {
			EncryptAndFill(scope)
			return
		}

		for f, info := range getDecryptedFields(scope) {
			key := strcase.ToSnake(f)
			field := reflectAttrs.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				continue
			}
			encrypted := Encrypt(field.Interface())
			reflectAttrs.SetMapIndex(reflect.ValueOf(getField(scope, info.encryptTo).db), reflect.ValueOf(encrypted))
			if !config.MigrateMode {
				reflectAttrs.SetMapIndex(reflect.ValueOf(key), reflect.Value{})
			}
		}

		scope.InstanceSet("gorm:update_interface", reflectAttrs.Interface())
	} else {
		EncryptAndFill(scope)
	}
}

func afterUpdate(scope *gorm.Scope) {
	if guard(scope) {
		return
	}
	CheckDecryptAndFill(scope)
}

func guard(scope *gorm.Scope) bool {
	_, skip := scope.DB().Get("secure_storage:skip")
	if skip {
		return true
	}
	return false
}
