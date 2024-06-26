package middleware

import (
	"errors"
	"fmt"
	"go/token"
	"time"
	"user-service/internal/module/user/repositories"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type Middleware struct {
	Repo repositories.Repositories
	Log  *otelzap.Logger
}

func (m *Middleware) VerifyBearerToken(ctx *fiber.Ctx) error {
	// get token from header
	auth := ctx.Get("Authorization")
	if auth == "" {
		m.Log.Ctx(ctx.Context()).Error("error get token from header")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// grab token
	token := auth[7:token.Pos(len(auth))]

	// decode token
	userID, err := decodeToken(token)
	if err != nil {
		m.Log.Ctx(ctx.Context()).Error(fmt.Sprintf("error decode token: %s", err.Error()))
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// validate user id
	result, err := m.Repo.FindUserByID(ctx.Context(), userID)
	if err != nil {
		m.Log.Ctx(ctx.Context()).Error(fmt.Sprintf("error find user by id: %s", err.Error()))
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// set user id to context
	ctx.Locals("userID", result.ID)

	return ctx.Next()
}

type CustomClaims struct {
	UserID    int   `json:"id"`
	ExpiredAt int64 `json:"exp"`
	jwt.StandardClaims
}

func decodeToken(jwtToken string) (int, error) {
	// Decode Token JWT
	token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod("HS256") != token.Method {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.ExpiredAt < time.Now().Unix() {
			return 0, errors.New("token expired")
		}
		return claims.UserID, nil
	} else {
		return 0, err
	}

}
