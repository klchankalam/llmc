package main

import (
	"db"
	"entity"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	rh "requesthandler"
)

func main() {
	// setup routes
	router := httprouter.New()
	router.POST("/orders", rh.HandleNewOrder)
	router.PATCH("/orders/:id", rh.HandleTakeOrder)
	router.GET("/orders", rh.HandleListOrder)

	// setup db
	log.Println("initializing DB...")
	db.InitDb()
	DB := db.GetDB()
	defer DB.Close()
	DB.AutoMigrate(&entity.Order{})
	log.Println("DB initialized")

	// start server
	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
}
