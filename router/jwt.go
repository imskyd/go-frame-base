package router

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

const (
	HeaderJwt = "Cookie"
)

func (srv *Service) createToken(userId int64, subId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id":     userId,
		"sub_id":      subId,
		"expire_time": time.Now().Unix() + int64(time.Second*86400*3), // Token过期时间，目前是24小时
	})
	pri := strings.Replace(srv.jwtConfig.PrivateKey, "||", "\n", -1)
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(pri))
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return tokenString, nil
}

func (srv *Service) createJwtToken(m *jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, m)
	pri := strings.Replace(srv.jwtConfig.PrivateKey, "||", "\n", -1)
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(pri))
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return tokenString, nil
}

func parse(tokenString string, key interface{}) (interface{}, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf(" Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return "", false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		fmt.Println("======pares:", err)
		return "", false
	}

}
