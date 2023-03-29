package secure_storage

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-web-kits/utils/project"
	"github.com/go-web-kits/utils/security"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func Encrypt(data interface{}) string {
	encrypted, err := security.Encrypt(Compress(data), config.Key)
	if err != nil {
		encrypted = "Encrypt Failed"
		if project.OnDevOrTest() {
			encrypted += ": " + err.Error()
		}
	}
	return encrypted
}

func Decrypt(data string) interface{} {
	decrypted, err := security.Decrypt(data, config.Key)
	if err != nil {
		decrypted = "Decrypt Failed"
		if project.OnDevOrTest() {
			decrypted += ": " + err.Error()
		}
	}

	unCompressed, err := UnCompress(decrypted)
	if err != nil {
		decrypted = "Decrypt Failed"
		if project.OnDevOrTest() {
			decrypted += ": " + err.Error()
		}
	}
	return unCompressed
}

// ========

// Compress("abc") => "abc"
// Compress(map[string]string{"a": "b"}) => "map[string]string##{\"a\":\"b\"}"
func Compress(data interface{}) string {
	var valString string
	t := reflect.TypeOf(data)

	switch t.Kind() {
	case reflect.TypeOf(postgres.Jsonb{}).Kind():
		panic("SecureStorage: unsupported type Jsonb")
		// bs, _ := d.(postgres.Jsonb).MarshalJSON()
		// valString = string(bs)
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		panic("SecureStorage: unsupported type Map/Slice/Struct")
	// bs, _ := json.Marshal(d)
	// valString = string(bs)
	case reflect.String:
		return data.(string)
	default:
		valString = fmt.Sprint(data)
	}

	return strings.Join([]string{t.String(), valString}, "##")
}

// UnCompress("map[string]string##{\"a\":\"b\"}") => map[string]interface{}{"a": "b"}
// UnCompress("map[int]int##{\"1\":2}") => map[string]interface{}{"1": 2.000000}
func UnCompress(compressed string) (interface{}, error) {
	info := strings.Split(compressed, "##")
	if len(info) == 1 { // string
		return info[0], nil
	}

	typeName, valString := info[0], info[1]
	var value interface{}
	var err error

	switch typeName {
	case "int":
		value, err = strconv.Atoi(valString)
	case "int32":
		value, err = strconv.ParseInt(valString, 10, 32)
	case "int64":
		value, err = strconv.ParseInt(valString, 10, 64)
	case "uint":
		value, err = strconv.Atoi(valString)
		value = uint(value.(int))
	case "uint8":
		value, err = strconv.Atoi(valString)
		value = byte(value.(int))
	case "uint64":
		value, err = strconv.ParseUint(valString, 10, 64)
	case "float32":
		value, err = strconv.ParseFloat(valString, 32)
	case "float64":
		value, err = strconv.ParseFloat(valString, 64)
	case "bool":
		if valString == "true" {
			value = true
		} else {
			value = false
		}
	default:
		value = valString
	}

	return value, err
}
