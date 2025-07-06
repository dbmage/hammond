package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CanInitializeSystem() (bool, error) {
	users, _ := GetAllUsers()
	if len(*users) != 0 {
		// db.MigrateClarkson("root:password@tcp(192.168.0.117:3306)/clarkson?charset=utf8mb4&parseTime=True&loc=Local")
		return false,
			fmt.Errorf("there are already users in the database. Migration can only be done on an empty database")
	}
	return true, nil
}

func CreateUser(user *User) error {
	tx := DB.Create(&user)
	return tx.Error
}

func UpdateUser(user *User) error {
	tx := DB.Omit(clause.Associations).Save(&user)
	return tx.Error
}

func FindOneUser(condition interface{}) (User, error) {

	var model User
	err := DB.Where(condition).First(&model).Error
	return model, err
}

func SetDisabledStatusForUser(userId uuid.UUID, isDisabled bool) error {
	//Cannot do this for admin
	tx := DB.Debug().Model(&User{}).Where("id= ? and role=?", userId, USER).Update("is_disabled", isDisabled)
	return tx.Error
}

func GetAllUsers() (*[]User, error) {

	sorting := "created_at desc"
	var users []User
	result := DB.Order(sorting).Find(&users)
	return &users, result.Error
}

func GetAllVehicles(sorting string) (*[]Vehicle, error) {
	if sorting == "" {
		sorting = "created_at desc"
	}
	var vehicles []Vehicle
	result := DB.Preload("Fillups", func(db *gorm.DB) *gorm.DB {
		return db.Order("fillups.date DESC")
	}).Preload("Expenses", func(db *gorm.DB) *gorm.DB {
		return db.Order("expenses.date DESC")
	}).Order(sorting).Find(&vehicles)
	return &vehicles, result.Error
}

func GetVehicleOwner(vehicleId uuid.UUID) (uuid.UUID, error) {
	var mapping UserVehicle

	tx := DB.Where("vehicle_id = ? AND is_owner = 1", vehicleId).First(&mapping)

	if tx.Error != nil {
		return uuid.UUID{}, tx.Error
	}
	return mapping.UserID, nil
}

func GetVehicleUsers(vehicleId uuid.UUID) (*[]UserVehicle, error) {
	var mapping []UserVehicle

	tx := DB.Debug().Preload("User").Where("vehicle_id = ?", vehicleId).Find(&mapping)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return &mapping, nil
}

func ShareVehicle(vehicleId, userId uuid.UUID) error {
	var mapping UserVehicle

	tx := DB.Where("vehicle_id = ? AND user_id = ?", vehicleId, userId).First(&mapping)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		newMapping := UserVehicle{
			UserID:    userId,
			VehicleID: vehicleId,
			IsOwner:   false,
		}
		tx = DB.Create(&newMapping)
		return tx.Error
	}
	return nil
}

func TransferVehicle(vehicleId, ownerId, newUserID uuid.UUID) error {

	tx := DB.Model(&UserVehicle{}).Where("vehicle_id = ? AND user_id = ?", vehicleId, ownerId).Update("is_owner", false)
	if tx.Error != nil {
		return tx.Error
	}
	tx = DB.Model(&UserVehicle{}).Where("vehicle_id = ? AND user_id = ?", vehicleId, newUserID).Update("is_owner", true)

	return tx.Error
}

func UnshareVehicle(vehicleId, userId uuid.UUID) error {
	var mapping UserVehicle

	tx := DB.Where("vehicle_id = ? AND user_id = ?", vehicleId, userId).First(&mapping)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	if mapping.IsOwner {
		return fmt.Errorf("cannot unshare owner")
	}
	result := DB.Where("id=?", mapping.ID).Delete(&UserVehicle{})
	return result.Error
}

func GetUserVehicles(id uuid.UUID) (*[]Vehicle, error) {
	var vehicles []Vehicle

	err := DB.
		Model(&Vehicle{}).
		Joins("JOIN user_vehicles ON user_vehicles.vehicle_id = vehicles.id").
		Where("user_vehicles.user_id = ?", id). // Make sure `id` is uuid.UUID
		Select("vehicles.*, user_vehicles.is_owner").
		Preload("Fillups", func(db *gorm.DB) *gorm.DB {
			return db.Order("fillups.date DESC")
		}).
		Preload("Expenses", func(db *gorm.DB) *gorm.DB {
			return db.Order("expenses.date DESC")
		}).
		Find(&vehicles).Error

	if err != nil {
		return nil, err
	}
	return &vehicles, nil
}

func GetUserById(id uuid.UUID) (*User, error) {
	var data User
	result := DB.Preload(clause.Associations).First(&data, "id=?", id)
	return &data, result.Error
}

func GetVehicleById(id uuid.UUID) (*Vehicle, error) {
	var vehicle Vehicle
	result := DB.Preload(clause.Associations).First(&vehicle, "id=?", id)
	return &vehicle, result.Error
}

func GetFillupById(id uuid.UUID) (*Fillup, error) {
	var obj Fillup
	result := DB.Preload(clause.Associations).First(&obj, "id=?", id)
	return &obj, result.Error
}

func GetFillupsByVehicleId(id uuid.UUID) (*[]Fillup, error) {
	var obj []Fillup
	result := DB.Preload(clause.Associations).Order("date desc").Find(&obj, &Fillup{VehicleID: id})
	return &obj, result.Error
}

func GetLatestFillupsByVehicleId(id uuid.UUID) (*Fillup, error) {
	var obj Fillup
	result := DB.Preload(clause.Associations).Order("date desc").First(&obj, &Fillup{VehicleID: id})
	return &obj, result.Error
}

func GetFillupsByVehicleIdSince(id uuid.UUID, since time.Time) (*[]Fillup, error) {
	var obj []Fillup
	result := DB.Where("date >= ? AND vehicle_id = ?", since, id).Preload(clause.Associations).Order("date desc").Find(&obj)
	return &obj, result.Error
}

func FindFillups(condition interface{}) (*[]Fillup, error) {

	var model []Fillup
	err := DB.Where(condition).Find(&model).Error
	return &model, err
}

func FindFillupsForDateRange(vehicleIds []uuid.UUID, start, end time.Time) (*[]Fillup, error) {

	var model []Fillup
	err := DB.Where("date <= ? AND date >= ? AND vehicle_id in ?", end, start, vehicleIds).Find(&model).Error
	return &model, err
}

func FindExpensesForDateRange(vehicleIds []uuid.UUID, start, end time.Time) (*[]Expense, error) {

	var model []Expense
	err := DB.Where("date <= ? AND date >= ? AND vehicle_id in ?", end, start, vehicleIds).Find(&model).Error
	return &model, err
}

func GetExpensesByVehicleId(id uuid.UUID) (*[]Expense, error) {
	var obj []Expense
	result := DB.Preload(clause.Associations).Order("date desc").Find(&obj, &Expense{VehicleID: id})
	return &obj, result.Error
}

func GetLatestExpenseByVehicleId(id uuid.UUID) (*Expense, error) {
	var obj Expense
	result := DB.Preload(clause.Associations).Order("date desc").First(&obj, &Expense{VehicleID: id})
	return &obj, result.Error
}

func GetExpenseById(id uuid.UUID) (*Expense, error) {
	var obj Expense
	result := DB.Preload(clause.Associations).First(&obj, "id=?", id)
	return &obj, result.Error
}

func DeleteVehicleById(id uuid.UUID) error {

	result := DB.Where("id=?", id).Delete(&Vehicle{})
	return result.Error
}

func DeleteFillupById(id uuid.UUID) error {

	result := DB.Where("id=?", id).Delete(&Fillup{})
	return result.Error
}

func DeleteExpenseById(id uuid.UUID) error {
	result := DB.Where("id=?", id).Delete(&Expense{})
	return result.Error
}

func DeleteFillupByVehicleId(id uuid.UUID) error {

	result := DB.Where("vehicle_id=?", id).Delete(&Fillup{})
	return result.Error
}

func DeleteExpenseByVehicleId(id uuid.UUID) error {
	result := DB.Where("vehicle_id=?", id).Delete(&Expense{})
	return result.Error
}

func GetAllQuickEntries(sorting string) (*[]QuickEntry, error) {
	if sorting == "" {
		sorting = "created_at desc"
	}
	var quickEntries []QuickEntry
	result := DB.Preload(clause.Associations).Order(sorting).Find(&quickEntries)
	return &quickEntries, result.Error
}

func GetQuickEntriesForUser(userId uuid.UUID, sorting string) (*[]QuickEntry, error) {
	if sorting == "" {
		sorting = "created_at desc"
	}
	var quickEntries []QuickEntry
	result := DB.Preload(clause.Associations).Where("user_id = ?", userId).Order(sorting).Find(&quickEntries)
	return &quickEntries, result.Error
}

func GetQuickEntryById(id uuid.UUID) (*QuickEntry, error) {
	var quickEntry QuickEntry
	result := DB.Preload(clause.Associations).First(&quickEntry, "id=?", id)
	return &quickEntry, result.Error
}

func DeleteQuickEntryById(id uuid.UUID) error {
	result := DB.Where("id=?", id).Delete(&QuickEntry{})
	return result.Error
}

func UpdateQuickEntry(entry *QuickEntry) error {
	return DB.Save(entry).Error
}

func SetQuickEntryAsProcessed(id uuid.UUID, processDate time.Time) error {
	result := DB.Model(QuickEntry{}).Where("id=?", id).Update("process_date", processDate)
	return result.Error
}

func GetAttachmentById(id uuid.UUID) (*Attachment, error) {
	var entry Attachment
	result := DB.Preload(clause.Associations).First(&entry, "id=?", id)
	return &entry, result.Error
}

func GetVehicleAttachments(vehicleId uuid.UUID) (*[]Attachment, error) {
	var attachments []Attachment
	vehicle, err := GetVehicleById(vehicleId)
	if err != nil {
		return nil, err
	}
	err = DB.Debug().Model(vehicle).Select("attachments.*,vehicle_attachments.title").Preload("User").Association("Attachments").Find(&attachments)
	if err != nil {
		return nil, err
	}
	return &attachments, nil
}

func GeAlertById(id uuid.UUID) (*VehicleAlert, error) {
	var alert VehicleAlert
	result := DB.Preload(clause.Associations).First(&alert, "id=?", id)
	return &alert, result.Error
}

func GetAlertOccurenceByAlertId(id uuid.UUID) (*[]AlertOccurance, error) {
	var alertOccurance []AlertOccurance
	result := DB.Preload(clause.Associations).Order("created_at desc").Find(&alertOccurance, "vehicle_alert_id=?", id)
	return &alertOccurance, result.Error
}

func GetUnprocessedAlertOccurances() (*[]AlertOccurance, error) {
	var alertOccurance []AlertOccurance
	result := DB.Preload(clause.Associations).Order("created_at desc").Find(&alertOccurance, "process_date is NULL")
	return &alertOccurance, result.Error
}

func MarkAlertOccuranceAsProcessed(id uuid.UUID, alertProcessType AlertType, date time.Time) error {
	tx := DB.Debug().Model(&AlertOccurance{}).Where("id= ?", id).
		Update("alert_process_type", alertProcessType).
		Update("process_date", date)
	return tx.Error

}

func UpdateSettings(setting *Setting) error {
	tx := DB.Save(&setting)
	return tx.Error
}

func GetOrCreateSetting() *Setting {
	var setting Setting
	result := DB.First(&setting)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		DB.Save(&Setting{})
		DB.First(&setting)
	}
	return &setting
}

func GetLock(name string) *JobLock {
	var jobLock JobLock
	result := DB.Where("name = ?", name).First(&jobLock)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return &JobLock{
			Name: name,
		}
	}
	return &jobLock
}

func Lock(name string, duration int) {
	jobLock := GetLock(name)
	if jobLock == nil {
		jobLock = &JobLock{
			Name: name,
		}
	}
	jobLock.Duration = duration
	jobLock.Date = time.Now()
	if jobLock.ID == uuid.Nil {
		DB.Create(&jobLock)
	} else {
		DB.Save(&jobLock)
	}
}

func Unlock(name string) {
	jobLock := GetLock(name)
	if jobLock == nil {
		return
	}
	jobLock.Duration = 0
	jobLock.Date = time.Time{}
	DB.Save(&jobLock)
}

func UnlockMissedJobs() {
	var jobLocks []JobLock

	result := DB.Find(&jobLocks)
	if result.Error != nil {
		return
	}
	for _, job := range jobLocks {
		if (job.Date.Equal(time.Time{})) {
			continue
		}
		var duration = time.Duration(job.Duration)
		d := job.Date.Add(time.Minute * duration)
		if d.Before(time.Now()) {
			fmt.Println(job.Name + " is unlocked")
			Unlock(job.Name)
		}
	}
}
