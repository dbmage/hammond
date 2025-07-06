package controllers

import (
	"errors"
	"net/http"

	"hammond/common"
	"hammond/models"
	"hammond/service"

	"github.com/gin-gonic/gin"
)

func RegisterVehicleController(router *gin.RouterGroup) {
	router.POST("/vehicles", createVehicle)
	router.GET("/vehicles", getAllVehicles)
	router.GET("/vehicles/:id", getVehicleById)
	router.PUT("/vehicles/:id", updateVehicle)
	router.DELETE("/vehicles/:id", deleteVehicle)
	router.GET("/vehicles/:id/stats", getVehicleStats)
	router.GET("/vehicles/:id/users", getVehicleUsers)
	router.POST("/vehicles/:id/users/:subId", shareVehicle)
	router.DELETE("/vehicles/:id/users/:subId", unshareVehicle)
	router.POST("/vehicles/:id/users/:subId/transfer", transferVehicle)

	router.GET("/me/vehicles", getMyVehicles)
	router.GET("/me/stats", getMystats)

	router.GET("/vehicles/:id/fillups", getFillupsByVehicleId)
	router.GET("/vehicles/:id/fuelSubTypes", getFuelSubTypesByVehicleId)
	router.POST("/vehicles/:id/fillups", createFillup)
	router.GET("/vehicles/:id/fillups/:subId", getFillupById)
	router.PUT("/vehicles/:id/fillups/:subId", updateFillup)
	router.DELETE("/vehicles/:id/fillups/:subId", deleteFillup)

	router.GET("/vehicles/:id/expenses", getExpensesByVehicleId)
	router.POST("/vehicles/:id/expenses", createExpense)
	router.GET("/vehicles/:id/expenses/:subId", getExpenseById)
	router.PUT("/vehicles/:id/expenses/:subId", updateExpense)
	router.DELETE("/vehicles/:id/expenses/:subId", deleteExpense)

	router.POST("/vehicles/:id/attachments", createVehicleAttachment)
	router.GET("/vehicles/:id/attachments", getVehicleAttachments)
}

func createVehicle(c *gin.Context) {
	var request models.CreateVehicleRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	vehicle, err := service.CreateVehicle(request, id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicle", err))
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}

func getVehicleById(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
			return
		}
		vehicle, err := service.GetVehicleById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
			return
		}
		c.JSON(http.StatusOK, vehicle)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func updateVehicle(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery
	var updateVehicleModel models.UpdateVehicleRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateVehicleModel); err == nil {
			id, err := common.ToUUID(searchByIdQuery.ID)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
				return
			}
			err = service.UpdateVehicle(id, updateVehicleModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getAllVehicles(c *gin.Context) {
	vehicles, err := service.GetAllVehicles()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
		return
	}
	c.JSON(200, vehicles)

}

func getMyVehicles(c *gin.Context) {
	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	vehicles, err := service.GetUserVehicles(id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getMyVehicles", err))
		return
	}
	c.JSON(200, vehicles)

}

func getFillupsByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupsByVehicleId", err))
			return
		}
		fillups, err := service.GetFillupsByVehicleId(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupsByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, fillups)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getFuelSubTypesByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFuelSubTypesByVehicleId", err))
			return
		}
		fuelSubtypes, err := service.GetDistinctFuelSubtypesForVehicle(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFuelSubTypesByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, fuelSubtypes)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getExpensesByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpensesByVehicleId", err))
			return
		}
		data, err := service.GetExpensesByVehicleId(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpensesByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createFillup(c *gin.Context) {
	var request models.CreateFillupRequest
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&request); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
			return
		}
		fillup, err := service.CreateFillup(request)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createFillup", err))
			return
		}
		c.JSON(http.StatusCreated, fillup)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createExpense(c *gin.Context) {
	var request models.CreateExpenseRequest
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&request); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
			return
		}
		expense, err := service.CreateExpense(request)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createExpense", err))
			return
		}
		c.JSON(http.StatusCreated, expense)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func updateExpense(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery
	var updateExpenseModel models.UpdateExpenseRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateExpenseModel); err == nil {
			id, err := common.ToUUID(searchByIdQuery.SubID)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("updateExpense", err))
				return
			}
			err = service.UpdateExpense(id, updateExpenseModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("updateExpense", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func updateFillup(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery
	var updateFillupModel models.UpdateFillupRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateFillupModel); err == nil {
			id, err := common.ToUUID(searchByIdQuery.SubID)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("updateFillup", err))
				return
			}
			err = service.UpdateFillup(id, updateFillupModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("updateFillup", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func deleteExpense(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteExpense", err))
			return
		}
		err = service.DeleteExpenseById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteExpense", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func deleteFillup(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteFillup", err))
			return
		}
		err = service.DeleteFillupById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteFillup", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getExpenseById(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpenseById", err))
			return
		}
		obj, err := service.GetExpenseById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpenseById", err))
			return
		}
		c.JSON(http.StatusOK, obj)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getFillupById(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupById", err))
			return
		}
		obj, err := service.GetFillupById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupById", err))
			return
		}
		c.JSON(http.StatusOK, obj)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createVehicleAttachment(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery
	var dataModel models.CreateVehicleAttachmentModel
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&dataModel); err == nil {
			id, err := common.ToUUID(searchByIdQuery.ID)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			vehicle, err := service.GetVehicleById(id)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			attachment, err := saveUploadedFile(c, "file")
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			err = service.CreateVehicleAttachment(vehicle.ID, attachment.ID, dataModel.Title)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleAttachments(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleAttachments", err))
			return
		}
		vehicle, err := service.GetVehicleById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleAttachments", err))
			return
		}

		attachments, err := service.GetVehicleAttachments(vehicle.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
			return
		}
		c.JSON(http.StatusOK, attachments)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleUsers(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleUsers", err))
			return
		}
		vehicle, err := service.GetVehicleById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleUsers", err))
			return
		}
		data, err := service.GetVehicleUsers(vehicle.ID)

		var model []models.UserVehicleSimpleModel

		for _, item := range *data {
			model = append(model, models.UserVehicleSimpleModel{
				ID:        item.ID,
				UserID:    item.UserID,
				VehicleID: item.VehicleID,
				IsOwner:   item.IsOwner,
				Name:      item.User.Name,
			})
		}

		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleUsers", err))
			return
		}

		c.JSON(http.StatusOK, model)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func deleteVehicle(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		id, err := common.ToUUID(c.MustGet("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
		}

		searchID, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteVehicle", err))
			return
		}
		canDelete, err := service.CanDeleteVehicle(searchID, id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteVehicle", err))
			return
		}
		if !canDelete {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteVehicle", errors.New("you are not allowed to delete this vehicle")))
			return
		}
		err = service.DeleteVehicle(searchID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func shareVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		subID, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}

		err = service.ShareVehicle(id, subID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func transferVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		id, err := common.ToUUID(c.MustGet("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
		}

		searchID, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("transferVehicle", err))
			return
		}
		subID, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("transferVehicle", err))
			return
		}
		err = service.TransferVehicle(searchID, id, subID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("transferVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func unshareVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("unShareVehicle", err))
			return
		}
		subID, err := common.ToUUID(searchByIdQuery.SubID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("unShareVehicle", err))
			return
		}
		err = service.UnshareVehicle(id, subID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("unShareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleStats(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleStats", err))
			return
		}
		vehicle, err := service.GetVehicleById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleStats", err))
			return
		}

		model := models.VehicleStatsModel{}

		c.JSON(http.StatusOK, model.SetStats(&vehicle.Fillups, &vehicle.Expenses))

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getMystats(c *gin.Context) {
	var model models.UserStatsQueryModel
	if err := c.ShouldBind(&model); err == nil {

		id, err := common.ToUUID(c.MustGet("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
		}
		stats, err := service.GetUserStats(id, model)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getMyVehicles", err))
			return
		}
		c.JSON(200, stats)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}

}
