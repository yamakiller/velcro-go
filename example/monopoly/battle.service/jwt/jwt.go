package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type MyClaims struct {
	Group string `json:"group"`
	UID   string `json:"uid"`
	Auth  int32  `json:"auth"` // 1 admin
	Max   int32  `json:"max"`  // 最大人数
	jwt.StandardClaims
}

func GenToken(secret string, group string, uid string, auth int32, max int32, TokenExpireDuration time.Duration) (string, error) {
	// 创建我们自己声明的数据
	c := MyClaims{
		group,
		uid,
		auth,
		max,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			IssuedAt:  time.Now().Unix(), // 当前时间
			NotBefore: time.Now().Unix(),
			Issuer:    "vercro_nat", // 签发人
		},
	}

	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 返回token
	return token.SignedString([]byte(secret))
}

func ParseToken(secret string, tokenString string) (*MyClaims, error) {
	mc := new(MyClaims)
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid { // 校验token
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
