package main

import (
	"db"
	"distancehelper"
	"entity"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	rh "requesthandler"
)

func main() {
	// setup db
	log.Println("initializing DB...")
	db.InitDb()
	DB := db.GetDB()
	defer DB.Close()
	DB.AutoMigrate(&entity.Order{})
	log.Println("DB initialized")

	dep := &rh.Dependencies{DB: DB, Map: &distancehelper.GMapReal{}}

	// setup routes
	router := httprouter.New()
	router.POST("/orders", dep.HandleNewOrder)
	router.PATCH("/orders/:id", dep.HandleTakeOrder)
	router.GET("/orders", dep.HandleListOrder)

	// start server
	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
}
