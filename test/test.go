package test

import (
	"github.com/go-web-kits/dbx"
	"github.com/go-web-kits/secure_storage"
	. "github.com/go-web-kits/testx"
	"github.com/go-web-kits/testx/factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type H = map[string]interface{}

type Card struct {
	ID              uint   `db:"id"`
	Number          string `encrypt_to:"EncryptedNumber" gorm:"-" db:"-"`
	EncryptedNumber string `decrypt_to:"Number"          db:"encrypted_number"`

	Price          float64 `encrypt_to:"EncryptedPrice" gorm:"-" db:"-"`
	EncryptedPrice string  `decrypt_to:"Price"          db:"encrypted_price"`

	// Jsonb          postgres.Jsonb `encrypt_to:"EncryptedJsonb" gorm:"-" db:"-"`
	// EncryptedJsonb string         `decrypt_to:"Jsonb"          db:"encrypted_jsonb"`

	TestMigrate  string `encrypt_to:"XTestMigrate" db:"test_migrate"`
	XTestMigrate string `decrypt_to:"TestMigrate"  db:"x_test_migrate"`

	Abc string `db:"abc"`

	secure_storage.Enable
}

var _ = Describe("SecureStorage", func() {
	var (
		card Card
		p    *MonkeyPatches
	)

	BeforeEach(func() {
		card = Card{Number: "abcdef"}
	})

	AfterEach(func() {
		CleanData(&Card{})
		p.Check()
	})

	When("Creating", func() {
		It("does successfully", func() {
			factory.Create(&card)
			Expect(card.Number).To(Equal("abcdef"))
			Expect(secure_storage.Encrypt("abcdef")).To(Equal("tmmjjG6IbGouWPOzu+Zblw=="))
			Expect(card.EncryptedNumber).To(Equal("tmmjjG6IbGouWPOzu+Zblw=="))

			reloaded := dbx.FindById(&Card{}, card.ID).Data.(*Card)
			Expect(reloaded.EncryptedNumber).To(Equal(card.EncryptedNumber))
			Expect(reloaded.Number).To(Equal(card.Number))

			card = Card{Price: 12.34}
			factory.Create(&card)
			Expect(card.Price).To(Equal(12.34))
			Expect(card.EncryptedPrice).To(Equal(secure_storage.Encrypt("float64##12.34")))
		})
	})

	When("Updating", func() {
		BeforeEach(func() {
			factory.Create(&card)
		})

		It("does successfully with `Update`", func() {
			card.Number = "jklmn"
			dbx.Update(&card)
			Expect(card.Number).To(Equal("jklmn"))
			Expect(card.EncryptedNumber).To(Equal("wvm2mg57yaCNOoEs7fhifg=="))

			reloaded := dbx.FindById(&Card{}, card.ID).Data.(*Card)
			Expect(reloaded.EncryptedNumber).To(Equal(card.EncryptedNumber))
			Expect(reloaded.Number).To(Equal(card.Number))
		})

		It("does successfully with `UpdateBy`", func() {
			// string
			dbx.UpdateBy(&card, H{"number": "123456789011111e32r32r32r"})
			Expect(card.Number).To(Equal("123456789011111e32r32r32r"))
			Expect(card.EncryptedNumber).To(Equal("fMg4M2G8Uhz3SZvDiP8XoCjJQrPxuTtbnf/NeMyFG6M="))

			reloaded := dbx.FindById(&Card{}, card.ID).Data.(*Card)
			Expect(reloaded.EncryptedNumber).To(Equal(card.EncryptedNumber))
			Expect(reloaded.Number).To(Equal(card.Number))

			// float
			card = Card{Price: 12.34}
			factory.Create(&card)
			dbx.UpdateBy(&card, H{"price": 34.56})
			Expect(card.Price).To(Equal(34.56))
			Expect(card.EncryptedPrice).To(Equal(secure_storage.Encrypt("float64##34.56")))

			reloaded = dbx.FindById(&Card{}, card.ID).Data.(*Card)
			Expect(reloaded.Price).To(Equal(card.Price))
			Expect(reloaded.EncryptedPrice).To(Equal(card.EncryptedPrice))
		})
	})

	When("DBx Querying", func() {
		var (
			c     Card
			cards []Card
		)

		BeforeEach(func() {
			c = Card{}
			cards = []Card{}
			factory.Create(&card)
		})

		It("does successfully with `Find`", func() {
			// string
			Expect(dbx.Find(&c, dbx.EQ{"number": card.Number})).To(HaveFound())
			Expect(c.Number).To(Equal(card.Number))
			Expect(c.EncryptedNumber).To(Equal(card.EncryptedNumber))

			// float
			card, c = Card{Price: 12.34}, Card{}
			factory.Create(&card)
			Expect(dbx.Find(&c, dbx.EQ{"number": card.Number})).To(HaveFound())
			Expect(c.Price).To(Equal(12.34))
			Expect(c.EncryptedPrice).To(Equal(secure_storage.Encrypt("float64##12.34")))
		})

		It("does successfully with `Chain Find`", func() {
			// TODO Failed currently
			// Expect(dbx.Conn().Where(dbx.EQ{"number": card.Number}).Find(&c).Error).To(Succeed())
			// Expect(c.Number).To(Equal(card.Number))
			// Expect(c.EncryptedNumber).To(Equal(card.EncryptedNumber))
		})

		It("does successfully with `Where`", func() {
			Expect(dbx.Where(&cards, nil)).To(HaveFound())
			Expect(cards[0].Number).To(Equal(card.Number))
			Expect(cards[0].EncryptedNumber).To(Equal(card.EncryptedNumber))
		})

		It("encrypts and updates automatically if encrypted filed not encoded", func() {
			Expect(dbx.UpdateBy(&Card{}, H{"encrypted_number": "abcdef"})).To(HaveAffected())
			Expect(dbx.Find(&c, nil)).To(HaveFound())
			Expect(c.Number).To(Equal(card.Number))
			Expect(c.EncryptedNumber).To(Equal(card.EncryptedNumber))
		})

		It("encrypts and updates automatically if encrypted filed is blank (single)", func() {
			p = IsExpectedToCall(secure_storage.HaveNotSecureStorage).AndReturn(true)
			Expect(dbx.UpdateBy(&Card{}, H{"test_migrate": "aaa"})).To(HaveAffected())
			p.Reset()
			Expect(dbx.Find(&c, nil)).To(HaveFound())
			Expect(c.TestMigrate).To(Equal("aaa"))
			Expect(c.XTestMigrate).To(Equal("GalUwfogftf+ZNC8hkRl3w=="))

			s := []string{}
			Expect(dbx.Model(&Card{}).Pluck("x_test_migrate", &s)).To(Succeed())
			Expect(s).To(ContainElement("GalUwfogftf+ZNC8hkRl3w=="))
		})

		It("encrypts and updates automatically if encrypted filed is blank (batch)", func() {
			p = IsExpectedToCall(secure_storage.HaveNotSecureStorage).AndReturn(true)
			Expect(dbx.UpdateBy(&Card{}, H{"test_migrate": "aaa"})).To(HaveAffected())
			p.Reset()
			Expect(dbx.Where(&cards, nil)).To(HaveFound())
			Expect(cards[0].TestMigrate).To(Equal("aaa"))
			Expect(cards[0].XTestMigrate).To(Equal("GalUwfogftf+ZNC8hkRl3w=="))

			s := []string{}
			Expect(dbx.Model(&Card{}).Pluck("x_test_migrate", &s)).To(Succeed())
			Expect(s).To(ContainElement("GalUwfogftf+ZNC8hkRl3w=="))
		})
	})

	Describe("Migrate Mode", func() {
		It("works", func() {
			card = Card{Number: "abcdef", TestMigrate: "test"}
			factory.Create(&card)
			card = Card{}
			Expect(dbx.FindBy(&card, dbx.H{"test_migrate": "test"})).To(HaveFound())
			Expect(card.TestMigrate).To(Equal("test"))
			Expect(card.XTestMigrate).To(Equal("JS6QHxD7Bby/4f6YanbfwA=="))

			s := []string{}
			Expect(dbx.Model(&Card{}).Pluck("test_migrate", &s)).To(Succeed())
			Expect(s).To(ContainElement("test"))

			cards := []Card{}
			Expect(dbx.Conn().DB.Where("id = ?", card.ID).Find(&cards).Error).NotTo(HaveOccurred())
			Expect(cards[0].TestMigrate).To(Equal("test"))
			Expect(cards[0].XTestMigrate).To(Equal("JS6QHxD7Bby/4f6YanbfwA=="))
		})
	})
})
