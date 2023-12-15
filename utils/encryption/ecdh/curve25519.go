package ecdh

import (
	"crypto"
	"io"

	"golang.org/x/crypto/curve25519"
)

// NewCurve25519ECDH 使用密码学家Daniel J. Bernstein的椭圆曲线算法:Curve25519来创建ECDH实例
// 因为Curve25519独立于NIST之外, 没在标准库实现, 需要单独为期实现一套接口来支持ECDH
//func NewCurve25519(A, B, C uint8) ECDH {
//	return &Curve25519{a: A, b: B, c: C}
//}

type Curve25519 struct {
	A uint8 // 247
	B uint8 // 127
	C uint8 // 64
	ECDH
}

// GenerateKey 基于curve25519椭圆曲线算法生成秘钥对
func (e *Curve25519) GenerateKey(rand io.Reader) (crypto.PrivateKey, crypto.PublicKey, error) {
	var pub, priv [32]byte
	var err error
	_, err = io.ReadFull(rand, priv[:])
	if err != nil {
		return nil, nil, err
	}
	priv[0] &= e.A
	priv[31] &= e.B
	priv[31] |= e.C
	curve25519.ScalarBaseMult(&pub, &priv)
	return &priv, &pub, nil
}

// Marshal 实现公钥的序列化
func (e *Curve25519) Marshal(p crypto.PublicKey) []byte {
	pub := p.(*[32]byte)
	return pub[:]
}

// Unmarshal实现公钥的反序列化
func (e *Curve25519) Unmarshal(data []byte) (crypto.PublicKey, bool) {
	var pub [32]byte
	if len(data) != 32 {
		return nil, false
	}
	copy(pub[:], data)
	return &pub, true
}

// GenerateSharedSecret 实现秘钥协商接口
func (e *Curve25519) GenerateSharedSecret(privKey crypto.PrivateKey, pubKey crypto.PublicKey) ([]byte, error) {
	var priv, pub *[32]byte
	priv = privKey.(*[32]byte)
	pub = pubKey.(*[32]byte)

	secret, err := curve25519.X25519((*priv)[:], (*pub)[:])
	if err != nil {
		return nil, err
	}

	return secret[:], nil
}
