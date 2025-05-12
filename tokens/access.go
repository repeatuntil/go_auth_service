package tokens

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
    UserAgent   string
	ClientIp	string
    jwt.RegisteredClaims
}

func GetSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func CreateToken(ip, userAgent, userId string) (string, error) {
	claims := JwtClaims{
		UserAgent: userAgent,
		ClientIp: ip,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userId,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return unsigned.SignedString(GetSecret())
}

func VerifyToken(tokenHeader string) (*jwt.Token, error) {
	tokenStr, ok := strings.CutPrefix(tokenHeader, "Bearer ")
	if !ok {
		return nil, fmt.Errorf("invalid authorization header")
	}

	claims := &JwtClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return GetSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
