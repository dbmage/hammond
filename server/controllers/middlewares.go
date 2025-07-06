package controllers

import (
	"net/http"
	"os"
	"strings"

	"hammond/common"
	"hammond/db"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Strips 'BEARER ' prefix from token string
func stripBearerPrefixFromTokenString(tok string) (string, error) {
	// Should be a bearer token
	if len(tok) > 6 && strings.ToUpper(tok[0:6]) == "BEARER " {
		return tok[7:], nil
	}
	return tok, nil
}

// Extract  token from Authorization header
// Uses PostExtractionFilter to strip "TOKEN " prefix from header
var AuthorizationHeaderExtractor = &request.PostExtractionFilter{
	Extractor: request.HeaderExtractor{"Authorization"},
	Filter:    stripBearerPrefixFromTokenString,
}

// Extractor for OAuth2 access tokens.  Looks in 'Authorization'
// header then 'access_token' argument for a token.
var MyAuth2Extractor = &request.MultiExtractor{
	AuthorizationHeaderExtractor,
	request.ArgumentExtractor{"access_token"},
}

// A helper to write user_id and user_model to the context
func UpdateContextUserModel(c *gin.Context, my_user_id uuid.UUID) {
	var myUserModel db.User
	if my_user_id != uuid.Nil {
		db.DB.First(&myUserModel, "id = ?", my_user_id)
	}
	c.Set("userId", my_user_id)
	c.Set("userModel", myUserModel)
}

// You can custom middlewares yourself as the doc: https://github.com/gin-gonic/gin#custom-middleware
//
//	r.Use(AuthMiddleware(true))
func AuthMiddleware(auto401 bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		UpdateContextUserModel(c, uuid.UUID{})
		token, err := request.ParseFromRequest(c.Request, MyAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			b := ([]byte(os.Getenv("JWT_SECRET")))
			return b, nil
		})
		if err != nil {
			if auto401 {
				_ = c.AbortWithError(http.StatusUnauthorized, err)
			}
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			id, err := common.ToUUID(claims["id"])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{})
			}
			UpdateContextUserModel(c, id)
		}
	}
}
