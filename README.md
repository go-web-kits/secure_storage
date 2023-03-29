# SecureStorage

Transparent Encryption Storage / 开箱即用的数据库字段自动加解密存储方案

须配合 GORM 使用（不必须使用 `dbx`）

## Features

1. 接近零配置
2. 无感知：基于 GORM 回调，在查询、创建、更新有关字段时自动加解密
3. 支持 string / int / uint / float / bool 类型的加解密，用法均一致
4. 与 ruby/secure_storage 使用完全一致的 AES256-CBC 加密，可以无缝迁移

## Setup

1. Configuration
```go
package initializer

import (
	"github.com/go-web-kits/dbx"
	. "github.com/go-web-kits/secure_storage"
	"github.com/go-web-kits/utils/settings"
)

func InitSecureStorage() {
	_ = Init(dbx.Conn().DB, Config{Key: settings.String("key")})
}
```

请注意 AES256 key 长度为 32 位

2. Definition
```go
type User struct {
	Email          string `encrypt_to:"EncryptedEmail" gorm:"-"`
	EncryptedEmail string `decrypt_to:"Email"          db:"encrypted_email"`

	secure_storage.Enable
}
```

## Usage

### Create

```go
user := User{Email: "user@example.com"}
dbx.Create(&user)
user.Email          // "user@example.com"
user.EncryptedEmail // "<encrypted>"
```

### Update

1. gorm `UpdateColumns` / dbx `UpdateBy`
```go
dbx.UpdateBy(&user, map[string]interface{}{"email": "changed@example.com"})
user.Email          // "changed@example.com"
user.EncryptedEmail // "<encrypted>"
```

2. gorm `Updates` / dbx `Update`
```go
user.Email = "changed@example.com"
dbx.Update(&user)
user.Email          // "changed@example.com"
user.EncryptedEmail // "<encrypted>"
```

Tip: `dbx` Definition Uniqueness 填写 `Account` 而不是 `EncryptedEmail`

### Query

```go
dbx.First(&user)
user.Email          // "user@example.com"
user.EncryptedEmail // "<encrypted>"
```

注意：以下示例（自动改变查询条件中的加密字段）仅限于 `dbx` 的 `EQ / IN` 两个 Conditioner
```go
dbx.Find(&user, dbx.EQ{"email": "user@example.com"})
user.Email          // "user@example.com"
user.EncryptedEmail // "<encrypted>"
```

另外：**base64 decode 不成功的时候，会认为该值未加密，将会【自动】更新为加密值**

### Migrate Data

1. 初始化配置中设置 `MigrateMode` 为 `true`, 去除 `gorm:"-"`
2. 跑 Migration 增加 加密列
3. 存量数据迁移，例如
    ```go
    total, err := dbx.Model(reflectx.New(model)).Count()
    if err != nil {
        panic(err)
    }
    totalPage := int(total) % 100
    
    for page := 0; page < totalPage; page++ {
        dbx.Where(reflectx.New(model), nil, dbx.With{Page: page, Rows: 100})
    }
    ```
4. 拿掉 `MigrateMode`，同时发版有关加密列数据库操作的逻辑
5. remove 非加密列