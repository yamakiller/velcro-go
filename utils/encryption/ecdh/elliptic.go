package ecdh

import (
	"crypto"
	"crypto/elliptic"
	"io"
	"math/big"
)

/*func NewEllipticECDH(curve elliptic.Curve) ECDH {
	return &Elliptic{
		curve: curve,
	}
}*/

type Elliptic struct {
	ECDH
	Curve elliptic.Curve
}

type EllipticPublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}
type EllipticPrivateKey struct {
	D []byte
}

// GenerateKey 基于标准库的NIST椭圆曲线算法生成秘钥对
func (e *Elliptic) GenerateKey(rand io.Reader) (crypto.PrivateKey, crypto.PublicKey, error) {
	var d []byte
	var x, y *big.Int
	var priv *EllipticPrivateKey
	var pub *EllipticPublicKey
	var err error
	d, x, y, err = elliptic.GenerateKey(e.Curve, rand)
	if err != nil {
		return nil, nil, err
	}
	priv = &EllipticPrivateKey{
		D: d,
	}
	pub = &EllipticPublicKey{
		Curve: e.Curve,
		X:     x,
		Y:     y,
	}
	return priv, pub, nil
}

// Marshal用于公钥的序列化
func (e *Elliptic) Marshal(p crypto.PublicKey) []byte {
	pub := p.(*EllipticPublicKey)
	return elliptic.Marshal(e.Curve, pub.X, pub.Y)
}

// Unmarshal用于公钥的反序列化
func (e *Elliptic) Unmarshal(data []byte) (crypto.PublicKey, bool) {
	var key *EllipticPublicKey
	var x, y *big.Int
	x, y = elliptic.Unmarshal(e.Curve, data)
	if x == nil || y == nil {
		return key, false
	}
	key = &EllipticPublicKey{
		Curve: e.Curve,
		X:     x,
		Y:     y,
	}
	return key, true
}

// GenerateSharedSecret 通过自己的私钥和对方的公钥协商一个共享密码
func (e *Elliptic) GenerateSharedSecret(privKey crypto.PrivateKey, pubKey crypto.PublicKey) ([]byte, error) {
	priv := privKey.(*EllipticPrivateKey)
	pub := pubKey.(*EllipticPublicKey)
	x, _ := e.Curve.ScalarMult(pub.X, pub.Y, priv.D)
	return x.Bytes(), nil
}
