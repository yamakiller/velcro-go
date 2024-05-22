package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type MyClaims struct {
	Group   string `json:"group"`
	UID     string `json:"uid"`
	jwt.StandardClaims
}

func GenToken(secret string, TokenExpireDuration time.Duration, group string, uid string) (string, error) {
	// 创建我们自己声明的数据
	c := MyClaims{
		group,
		uid,
		jwt.StandardClaims{
			IssuedAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:   "vercro_nat",                               // 签发人
		},
	}

	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 返回token
	return token.SignedString(secret)
}

func ParseToken(secret string, tokenString string) (*MyClaims, error) {
	mc := new(MyClaims)
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i interface{}, err error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid { // 校验token
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
