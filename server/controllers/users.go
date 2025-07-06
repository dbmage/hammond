package controllers

import (
	"net/http"

	"hammond/common"
	"hammond/db"
	"hammond/models"
	"hammond/service"

	"github.com/gin-gonic/gin"
)

func RegisterUserController(router *gin.RouterGroup) {
	router.GET("/users", allUsers)
	router.POST("/users/:id/enable", ShouldBeAdmin(), enableUser)
	router.POST("/users/:id/disable", ShouldBeAdmin(), disableUser)
}

func allUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, users)

}

func enableUser(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		err = service.SetDisabledStatusForUser(id, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}

}

func disableUser(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		err = service.SetDisabledStatusForUser(id, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}

}
