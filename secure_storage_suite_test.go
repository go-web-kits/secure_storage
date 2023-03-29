package secure_storage_test

import (
	"fmt"
	"testing"

	"github.com/go-web-kits/dbx"
	. "github.com/go-web-kits/secure_storage"
	. "github.com/go-web-kits/secure_storage/test"
	. "github.com/go-web-kits/testx"
	"github.com/go-web-kits/utils/project"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSecureStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecureStorage Suite")
}

var _ = BeforeSuite(func() {
	dbx.Client, _ = getPostgresDB()
	Boot().Migrate(&Card{})
	_ = Init(dbx.Conn().DB, Config{Key: "3b9c54223eae794396a7f3863623f688"})
})

var _ = AfterSuite(func() {
	ShutApp()
})

// ======

// TODO: extra
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"user_name"`
	Database     string `json:"database"`
	TestDatabase string `json:"test_database"`
	Password     string `json:"password"`
	DisableSSL   bool   `json:"disable_ssl"`
}

func (c *DatabaseConfig) ConnectString() string {
	db := c.Database
	if project.OnTest() {
		db = c.TestDatabase
	}
	fmt.Println("* Connecting Database\n    using `" + db + "`")

	s := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s",
		c.Host, c.Port, c.Username, db, c.Password,
	)
	if c.DisableSSL {
		s += " sslmode=disable"
	}

	return s
}

func getPostgresDB() (db *gorm.DB, err error) {
	c := DatabaseConfig{
		Host: "localhost", Port: 5432, Database: "firmware_services_test",
		Username: "postgres", Password: "123456", DisableSSL: true,
	}
	db, err = gorm.Open("postgres", c.ConnectString())
	if err != nil {
		fmt.Println("    ERROR: " + err.Error())
		fmt.Println("    Host: " + c.Host)
		return nil, err
	}
	fmt.Println("    Connected!")
	return db, err
}
