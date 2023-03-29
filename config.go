package secure_storage

import (
	"github.com/go-web-kits/utils"
	"github.com/jinzhu/gorm"
)

type Config struct {
	Key         string
	MigrateMode bool

	models map[string]fields
}

type fields = map[string]field

type field struct {
	encrypted bool
	encryptTo string
	decryptTo string
	db        string
}

var config Config

func getFields(scope *gorm.Scope) fields {
	if config.Key == "" {
		panic("SecureStorage: Not Configured!")
	}

	model := utils.TypeNameOf(scope.Value)
	if config.models[model] != nil {
		return config.models[model]
	}

	if config.models == nil {
		config.models = map[string]fields{}
	}
	config.models[model] = fields{}

	for _, f := range scope.Fields() {
		encryptTo := f.Tag.Get("encrypt_to")
		decryptTo := f.Tag.Get("decrypt_to")
		if encryptTo != "" || decryptTo != "" {
			// if f.Field.Kind() != reflect.String {
			// 	panic("SecureStorage: Field `" + f.Name + "` must be a string")
			// }
			config.models[model][f.Name] = field{
				encrypted: decryptTo != "",
				encryptTo: encryptTo,
				decryptTo: decryptTo,
				db:        f.Tag.Get("db"),
			}
		}
	}

	return config.models[model]
}

func getEncryptedFields(scope *gorm.Scope) fields {
	result := fields{}
	for name, info := range getFields(scope) {
		if info.encrypted {
			result[name] = info
		}
	}
	return result
}

func getDecryptedFields(scope *gorm.Scope) fields {
	result := fields{}
	for name, info := range getFields(scope) {
		if !info.encrypted {
			result[name] = info
		}
	}
	return result
}

func getField(scope *gorm.Scope, fieldName string) field {
	f, ok := getFields(scope)[fieldName]
	if !ok {
		panic("SecureStorage: Check your Tags")
	}
	return f
}
