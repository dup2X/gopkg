# dcrypt
> 简单的AES加解密组件，主要用于auth-key的生成

## how to use
```
    key := "test_key"
    plainText := "example123/.,"
	cipherText, err := Encrypt(plainText, key)
    // 密文cipherText = "dcrypto-ab940f8d0abbcc20c9216d163bf6e5500bc22d5cdd1ee6a5a3ac705b2e"
	res, err := Decrypt(cipherText, key)
```

## bench_mark
```
BenchmarkEncrypt-4        500000          2577 ns/op
BenchmarkDecrypt-4       1000000          1394 ns/op
```
