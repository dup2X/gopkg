package dcrypto

import (
	"testing"
)

func TestCrypto(t *testing.T) {
	var (
		key       = "test_key"
		plainText = "example123/.,"
	)
	cipherText, err := Encrypt(plainText, key)
	if err != nil {
		t.Fatal(err.Error())
	}
	res, err := Decrypt(cipherText, key)
	if err != nil {
		t.FailNow()
	}
	if res != plainText {
		t.FailNow()
	}
}

func BenchmarkEncrypt(b *testing.B) {
	var (
		key       = "test_key"
		plainText = "example123/.,"
	)
	for i := 0; i < b.N; i++ {
		Encrypt(plainText, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	var (
		key       = "test_key"
		plainText = "example123/.,"
	)
	cipherText, _ := Encrypt(plainText, key)
	for i := 0; i < b.N; i++ {
		Decrypt(cipherText, key)
	}
}
