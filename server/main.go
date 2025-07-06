package main

import (
	"fmt"
	"log"
	"os"

	"hammond/controllers"
	"hammond/db"
	"hammond/service"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var err error
	db.DB, err = db.Init()
	if err != nil {
		log.Fatal("unable to initialise the database: ", err)
	}

	db.Migrate()

	r := gin.Default()
	r.Use(setupSettings())
	r.Use(gin.Recovery())
	r.Use(location.Default())
	r.Use(static.Serve("/", static.LocalFile("./dist", true)))
	r.NoRoute(func(c *gin.Context) {
		c.File("dist/index.html")
	})
	router := r.Group("/api")

	dataPath := os.Getenv("DATA")

	router.Static("/assets/", dataPath)

	controllers.RegisterAnonController(router)
	controllers.RegisterAnonMasterConroller(router)
	controllers.RegisterSetupController(router)

	router.Use(controllers.AuthMiddleware(true))
	controllers.RegisterUserController(router)
	controllers.RegisterMastersController(router)
	controllers.RegisterAuthController(router)
	controllers.RegisterVehicleController(router)
	controllers.RegisterFilesController(router)
	controllers.RegisteImportController(router)
	controllers.RegisterReportsController(router)

	go assetEnv()
	go intiCron()

	err = r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		log.Fatal("unable to start the server", err)
	}
}

func setupSettings() gin.HandlerFunc {
	return func(c *gin.Context) {

		setting := db.GetOrCreateSetting()
		c.Set("setting", setting)
		c.Writer.Header().Set("X-Clacks-Overhead", "GNU Terry Pratchett")

		c.Next()
	}
}

func intiCron() {

	err := gocron.Every(2).Days().Do(service.CreateBackup)
	if err != nil {
		fmt.Println("failed to setup cron job", err)
	}

	<-gocron.Start()
}

func assetEnv() {
	log.Println("Config Dir: ", os.Getenv("CONFIG"))
	log.Println("Assets Dir: ", os.Getenv("DATA"))
}
