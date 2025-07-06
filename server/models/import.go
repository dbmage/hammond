package models

import "github.com/google/uuid"

type ImportData struct {
	Data      []ImportFillup `json:"data" binding:"required"`
	VehicleId uuid.UUID      `gorm:"type:uuid" json:"vehicleId" binding:"required"`
	TimeZone  string         `json:"timezone" binding:"required"`
}

type ImportFillup struct {
	VehicleID       uuid.UUID `gorm:"type:uuid" json:"vehicleId"`
	FuelQuantity    float32   `json:"fuelQuantity"`
	PerUnitPrice    float32   `json:"perUnitPrice"`
	TotalAmount     float32   `json:"totalAmount"`
	OdoReading      int       `json:"odoReading"`
	IsTankFull      *bool     `json:"isTankFull"`
	HasMissedFillup *bool     `json:"hasMissedFillup"`
	Comments        string    `json:"comments"`
	FillingStation  string    `json:"fillingStation"`
	UserID          uuid.UUID `gorm:"type:uuid" json:"userId"`
	Date            string    `json:"date"`
	FuelSubType     string    `json:"fuelSubType"`
}
