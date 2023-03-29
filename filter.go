package secure_storage

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
)

func FilterMap(db *gorm.DB, values map[string]interface{}) map[string]interface{} {
	if config.Key == "" || db.Value == nil {
		return values
	}

	scope := db.NewScope(db.Value)
	for f, info := range getDecryptedFields(scope) {
		key := strcase.ToSnake(f)
		if values[key] == nil {
			continue
		}

		switch v := values[key].(type) {
		case string:
			values[getField(scope, info.encryptTo).db] = Encrypt(v)
		case []string:
			slice := []string{}
			for _, _v := range v {
				slice = append(slice, Encrypt(_v))
			}
			values[getField(scope, info.encryptTo).db] = slice
		default:
			fmt.Println("secure_storage.FilterMap: cannot process the value type")
			return values
		}
		delete(values, key)
	}

	return values
}
