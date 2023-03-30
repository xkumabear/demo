package public

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"time"
)


type MyCustomClaims struct {
	Foo string `json:"foo"` //自定义字段
	jwt.StandardClaims
}

func JwtDecode(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtSignKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
		if claims.ExpiresAt < time.Now().Unix() {
			return nil, errors.New("request expired")
		}
		return claims, nil
	} else {
		return nil, errors.New("token is not jwt.standard...")
	}
}

func JwtEncode(claims jwt.StandardClaims) (string, error) {
	mySigningKey := []byte(JwtSignKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}
