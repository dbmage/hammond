package controllers

import (
	"net/http"
	"strconv"

	"hammond/common"
	"hammond/models"
	"hammond/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisteImportController(router *gin.RouterGroup) {
	router.POST("/import/fuelly", fuellyImport)
	router.POST("/import/drivvo", drivvoImport)
	router.POST("/import/generic", genericImport)
}

func fuellyImport(c *gin.Context) {
	bytes, err := getFileBytes(c, "file")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	errors := service.FuellyImport(bytes, id)
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errors})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func drivvoImport(c *gin.Context) {
	bytes, err := getFileBytes(c, "file")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}
	vehicleIdString := c.PostForm("vehicleID")
	if vehicleIdString == "" {
		c.JSON(http.StatusUnprocessableEntity, "Missing Vehicle ID")
		return
	}

	vehicleId, err := uuid.Parse(vehicleIdString)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Missing Vehicle ID")
		return
	}
	importLocation, err := strconv.ParseBool(c.PostForm("importLocation"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Please include importLocation option.")
		return
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	errors := service.DrivvoImport(bytes, id, vehicleId, importLocation)
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errors})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func genericImport(c *gin.Context) {
	var json models.ImportData
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if json.VehicleId == uuid.Nil {
		c.JSON(http.StatusUnprocessableEntity, "Missing Vehicle ID")
		return
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	errors := service.GenericImport(json, id)
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errors})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
