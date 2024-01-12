package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/yamakiller/velcro-go/utils/encryption"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
)

func main() {
	e := &ecdh.Curve25519{A: 248, B: 127, C: 64}
	_, pub1, _ := e.GenerateKey(rand.Reader)
	pub1Byte := e.Marshal(pub1)

	fmt.Printf("公钥序列化Byte码:")
	for _, b := range pub1Byte {
		fmt.Printf("%02X ", b)
	}
	fmt.Printf("\n")
	fmt.Printf("公钥序发送到客户端Base64:%s\n", base64.StdEncoding.EncodeToString(pub1Byte))
	prv2, _, _ := e.GenerateKey(rand.Reader)
	//pub2Byte := e.Marshal(pub2)
	fmt.Printf("客户端私钥:%v\n", prv2)

	secret, _ := e.GenerateSharedSecret(prv2, pub1)
	fmt.Printf("客户端生成的共享密钥Byte码:")
	for _, b := range secret {
		fmt.Printf("%02X ", b)
	}
	fmt.Printf("\n")

	source := "1234567890"
	fmt.Printf("待加密数据:%s\n", source)

	encry, _ := encryption.AesEncryptByGCM([]byte(source), secret)
	fmt.Printf("加密后Byte码:")
	for _, b := range encry {
		fmt.Printf("%02X ", b)
	}
	fmt.Printf("\n")
}
