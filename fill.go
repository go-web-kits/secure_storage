package secure_storage

import (
	"fmt"
	"reflect"

	"github.com/go-web-kits/utils"
	"github.com/go-web-kits/utils/security"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
)

func CheckDecryptAndFill(scope *gorm.Scope) {
	indirectValue := scope.IndirectValue()
	if indirectValue.Kind() == reflect.Slice {
		for i := 0; i < indirectValue.Len(); i++ {
			itemCheckDecryptAndFill(scope, indirectValue.Index(i))
		}
	}

	if HaveNotSecureStorage(scope) {
		return
	}

	unEncrypted := map[string]string{}
	for field, info := range getEncryptedFields(scope) {
		reflectEncryptedField, ok := scope.FieldByName(field)
		if !ok {
			continue
		}

		var decrypted interface{}
		encrypted := reflectEncryptedField.Field.Interface()
		encryptedString, isString := encrypted.(string)
		// 加密字段为空，未加密字段不为空
		if isString && encryptedString == "" {
			decryptField, ok := scope.FieldByName(info.decryptTo)
			if !ok {
				continue
			}
			if decrypted = decryptField.Field.Interface(); decrypted == "" {
				continue
			}
			encryptedString = Encrypt(decrypted)
			_ = reflectEncryptedField.Set(encryptedString)
			unEncrypted[info.db] = encryptedString
			continue
		}

		// 加密字段不为空且非 base64
		if !isString || (isString && security.NotEncoded(encryptedString)) {
			decrypted = encrypted
			unEncrypted[info.db] = Encrypt(encrypted)
		} else {
			// 加密字段不为空且 base64，可以认为是已加密的正常值
			decrypted = Decrypt(encryptedString)
		}
		err := scope.SetColumn(info.decryptTo, decrypted)
		if err != nil {
			_, _ = fmt.Println(errors.Wrapf(err, "SecureStorage CheckDecryptAndFill"))
		}
	}

	if len(unEncrypted) > 0 && scope.PrimaryKeyValue() != uint(0) {
		_ = scope.DB().InstantSet("secure_storage:skip", true).UpdateColumns(unEncrypted)
	}
}

func EncryptAndFill(scope *gorm.Scope) {
	if HaveNotSecureStorage(scope) {
		return
	}

	for field, info := range getDecryptedFields(scope) {
		reflectDecryptedField, ok := scope.FieldByName(field)
		if !ok {
			continue
		}
		decrypted := reflectDecryptedField.Field.Interface()
		err := scope.SetColumn(info.encryptTo, Encrypt(decrypted))
		if err != nil {
			_, _ = fmt.Println(errors.Wrapf(err, "SecureStorage EncryptAndFill"))
		}
	}
}

// ====

func itemCheckDecryptAndFill(s *gorm.Scope, item reflect.Value) {
	indirectObj := reflect.Indirect(item)
	scope := s.New(indirectObj.Interface())
	if HaveNotSecureStorage(scope) {
		return
	}

	unEncrypted := map[string]string{}
	for field, info := range getEncryptedFields(scope) {
		reflectEncryptedField := indirectObj.FieldByName(field)
		if !reflectEncryptedField.IsValid() {
			continue
		}

		var decrypted interface{}
		encrypted := reflectEncryptedField.Interface()
		encryptedString, isString := encrypted.(string)
		if isString && encryptedString == "" {
			decryptField := indirectObj.FieldByName(info.decryptTo)
			if !decryptField.IsValid() {
				continue
			}
			if decrypted = decryptField.Interface(); decrypted == "" {
				continue
			}
			encryptedString = Encrypt(decrypted)
			reflectEncryptedField.Set(reflect.ValueOf(encryptedString))
			unEncrypted[info.db] = encryptedString
			continue
		}

		if !isString || (isString && security.NotEncoded(encryptedString)) {
			decrypted = encrypted
			unEncrypted[info.db] = Encrypt(encrypted)
		} else {
			decrypted = Decrypt(encryptedString)
		}

		indirectObj.FieldByName(info.decryptTo).Set(reflect.ValueOf(decrypted))
	}

	if len(unEncrypted) > 0 {
		i := item.Interface()
		// TODO
		// _ = scope.DB().InstantSet("secure_storage:skip", true).UpdateColumns(unEncrypted)
		_ = s.NewDB().Model(&i).Where("id = ?", scope.PrimaryKeyValue()).
			Table(inflection.Plural(strcase.ToSnake(utils.TypeNameOf(i)))).
			InstantSet("secure_storage:skip", true).UpdateColumns(unEncrypted)
	}
}
