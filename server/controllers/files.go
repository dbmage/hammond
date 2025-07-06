package controllers

import (
	"errors"
	"io"
	"net/http"
	"os"

	"hammond/common"
	"hammond/db"
	"hammond/models"
	"hammond/service"

	"github.com/gin-gonic/gin"
)

func RegisterFilesController(router *gin.RouterGroup) {
	router.POST("/upload", uploadFile)
	router.POST("/quickEntries", createQuickEntry)
	router.GET("/quickEntries", getAllQuickEntries)
	router.GET("/me/quickEntries", getMyQuickEntries)
	router.GET("/quickEntries/:id", getQuickEntryById)
	router.POST("/quickEntries/:id/process", setQuickEntryAsProcessed)
	router.DELETE("/quickEntries/:id", deleteQuickEntryById)

	router.GET("/attachments/:id/file", getAttachmentFile)
}

func createQuickEntry(c *gin.Context) {
	var request models.CreateQuickEntryModel
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	attachment, err := saveUploadedFile(c, "file")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError("createQuickEntry", errors.New("userId is not a valid uuid")))
		return
	}
	quickEntry, err := service.CreateQuickEntry(request, attachment.ID, id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("createQuickEntry", err))
		return
	}
	c.JSON(http.StatusCreated, quickEntry)
}

func getAllQuickEntries(c *gin.Context) {
	quickEntries, err := service.GetAllQuickEntries("")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getAllQuickEntries", err))
		return
	}
	c.JSON(http.StatusOK, quickEntries)
}

func getMyQuickEntries(c *gin.Context) {
	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
	quickEntries, err := service.GetQuickEntriesForUser(id, "")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getMyQuickEntries", err))
		return
	}
	c.JSON(http.StatusOK, quickEntries)
}

func getQuickEntryById(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getQuickEntryById", err))
			return
		}
		quickEntry, err := service.GetQuickEntryById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getQuickEntryById", err))
			return
		}
		c.JSON(http.StatusOK, quickEntry)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func deleteQuickEntryById(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteQuickEntryById", err))
			return
		}
		err = service.DeleteQuickEntryById(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("deleteQuickEntryById", err))
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func setQuickEntryAsProcessed(c *gin.Context) {
	var searchByIdQuery models.SearchByIDQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		id, err := common.ToUUID(searchByIdQuery.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("setQuickEntryAsProcessed", err))
			return
		}
		err = service.SetQuickEntryAsProcessed(id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("setQuickEntryAsProcessed", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func uploadFile(c *gin.Context) {
	attachment, err := saveMultipleUploadedFile(c, "file")
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, attachment)
	}
}

func getAttachmentFile(c *gin.Context) {
	var query models.SearchByIDQuery

	// Bind URI param
	if err := c.ShouldBindUri(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid or missing ID",
			"error":   err.Error(),
		},
		)
		return
	}

	id, err := common.ToUUID(query.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Fetch attachment
	attachment, err := db.GetAttachmentById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(attachment.Path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	// Serve file
	c.File(attachment.Path)
}

func getFileBytes(c *gin.Context, fileVariable string) ([]byte, error) {
	if fileVariable == "" {
		fileVariable = "file"
	}
	formFile, err := c.FormFile(fileVariable)
	if err != nil {
		return nil, err
	}
	openedFile, _ := formFile.Open()
	return io.ReadAll(openedFile)
}

func saveUploadedFile(c *gin.Context, fileVariable string) (*db.Attachment, error) {
	if fileVariable == "" {
		fileVariable = "file"
	}
	file, err := c.FormFile(fileVariable)
	if err != nil {
		return nil, err
	}
	filePath := service.GetFilePath(file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return nil, err
	}

	id, err := common.ToUUID(c.MustGet("userId"))
	if err != nil {
		return nil, errors.New("unable to parse user ID")
	}
	return service.CreateAttachment(filePath, file.Filename, file.Size, file.Header.Get("Content-Type"), id)
}

func saveMultipleUploadedFile(c *gin.Context, fileVariable string) ([]*db.Attachment, error) {
	if fileVariable == "" {
		fileVariable = "files"
	}
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[fileVariable]
	var toReturn []*db.Attachment
	for _, file := range files {
		filePath := service.GetFilePath(file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			return nil, err
		}
		id, err := common.ToUUID(c.MustGet("userId"))
		if err != nil {
			return nil, errors.New("unable to parse user ID")
		}
		attachment, err := service.CreateAttachment(filePath, file.Filename, file.Size, file.Header.Get("Content-Type"), id)
		if err != nil {
			return nil, err
		}

		toReturn = append(toReturn, attachment)
	}
	return toReturn, nil
}
