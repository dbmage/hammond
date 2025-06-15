package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type localMigration struct {
	Name  string
	Query string
}

var migrations = []localMigration{
	{
		Name:  "2021_06_24_04_42_SetUserDisabledFalse",
		Query: "update users set is_disabled=0",
	},
	{
		Name:  "2021_02_07_00_09_LowerCaseEmails",
		Query: "update users set email=lower(email)",
	},
}

func RunMigrations() {
	for _, mig := range migrations {
		err := ExecuteAndSaveMigration(mig.Name, mig.Query)
		if err != nil {
			fmt.Printf("migration failed for '%s' due to error: %s\n", mig.Name, err)
			return
		}
	}
}

func ExecuteAndSaveMigration(name string, query string) error {
	var migration Migration
	result := DB.Where("name=?", name).First(&migration)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(query)
		result = DB.Debug().Exec(query)
		if result.Error == nil {
			DB.Save(&Migration{
				Date: time.Now(),
				Name: name,
			})
		}
		return result.Error
	}

	return nil
}
