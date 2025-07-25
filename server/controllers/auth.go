package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"hammond/common"
	"hammond/db"
	"hammond/models"
	"hammond/service"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterAnonController(router *gin.RouterGroup) {
	router.POST("/login", userLogin)
	router.POST("/auth/initialize", initializeSystem)

}

func RegisterAuthController(router *gin.RouterGroup) {

	router.POST("/refresh", refresh)
	router.GET("/me", me)
	router.POST("/register", ShouldBeAdmin(), userRegister)
	router.POST("/changePassword", changePassword)

}

func ShouldBeAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		model := c.MustGet("userModel").(db.User)
		if model.Role != db.ADMIN {
			c.JSON(http.StatusUnauthorized, gin.H{})
		} else {
			c.Next()
		}
	}
}

func me(c *gin.Context) {
	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}

	user, err := service.GetUserById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	c.JSON(http.StatusOK, user)
}

func userRegister(c *gin.Context) {
	var registerRequest models.RegisterRequest
	if err := c.ShouldBind(&registerRequest); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}

	if err := service.CreateUser(&registerRequest, *registerRequest.Role); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func initializeSystem(c *gin.Context) {

	canInitialize, err := service.CanInitializeSystem()
	if !canInitialize {
		c.JSON(http.StatusUnprocessableEntity, err)
	}

	var registerRequest models.RegisterRequest
	if err := c.ShouldBind(&registerRequest); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}

	_ = service.UpdateSettings(registerRequest.Currency, *registerRequest.DistanceUnit)

	if err := service.CreateUser(&registerRequest, db.ADMIN); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("initializeSystem", err))
		return
	}

	_ = service.UpdateSettings(registerRequest.Currency, *registerRequest.DistanceUnit)

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func userLogin(c *gin.Context) {
	var loginRequest models.LoginRequest
	if err := c.ShouldBind(&loginRequest); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	user, err := db.FindOneUser(&db.User{Email: strings.ToLower(loginRequest.Email)})

	if err != nil {
		c.JSON(http.StatusForbidden, common.NewError("login", errors.New("not Registered email or invalid password")))
		return
	}

	if user.CheckPassword(loginRequest.Password) != nil {
		c.JSON(http.StatusForbidden, common.NewError("login", errors.New("not Registered email or invalid password")))
		return
	}

	if user.IsDisabled {
		c.JSON(http.StatusForbidden, common.NewError("login", errors.New("your user has been disabled by the admin. Please contact them to get it re-enabled")))
		return
	}
	UpdateContextUserModel(c, user.ID)
	token, refreshToken := common.GenToken(user.ID, user.Role)
	response := models.LoginResponse{
		Name:         user.Name,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		Role:         user.RoleDetail().Key,
	}
	c.JSON(http.StatusOK, response)
}

func refresh(c *gin.Context) {
	type tokenReqBody struct {
		RefreshToken string `json:"refreshToken"`
	}
	tokenReq := tokenReqBody{}
	if err := c.Bind(&tokenReq); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	token, _ := jwt.Parse(tokenReq.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Get the user record from database or
		// run through your business logic to verify if the user can log in
		id, err := common.ToUUID(claims["id"])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
		}
		user, err := service.GetUserById(id)
		if err == nil {

			token, refreshToken := common.GenToken(user.ID, user.Role)

			response := models.LoginResponse{
				Name:         user.Name,
				Email:        user.Email,
				Token:        token,
				RefreshToken: refreshToken,
				Role:         user.RoleDetail().Key,
			}
			c.JSON(http.StatusOK, response)
		} else {

			c.JSON(http.StatusUnauthorized, gin.H{})
		}
	} else {

		c.JSON(http.StatusUnauthorized, gin.H{})
	}
}

func changePassword(c *gin.Context) {
	var request models.ChangePasswordRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}

	value := c.GetString("userId")
	id, err := uuid.Parse(value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.NewError("changePassword", errors.New("unable to change the password")))
		return
	}

	user, err := service.GetUserById(id)
	if err != nil {
		c.JSON(http.StatusForbidden, common.NewError("changePassword", errors.New("not Registered email or invalid password")))
		return
	}

	if user.CheckPassword(request.OldPassword) != nil {
		c.JSON(http.StatusForbidden, common.NewError("changePassword", errors.New("incorrect old password")))
		return
	}

	if err = user.SetPassword(request.NewPassword); err != nil {
		fmt.Println("error setting password: ", err)
		c.JSON(http.StatusInternalServerError, common.NewError("changePassword", errors.New("error setting password")))
	}

	success, _ := service.UpdatePassword(user.ID, request.NewPassword)
	c.JSON(http.StatusOK, success)
}
