package main

import (
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

var (
	httpService *gin.Engine
)

func Server() {
	httpService = gin.Default()
	httpService.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	SetRoutes()

	httpService.Run(":8080")
}

func SetRoutes() {
  httpService.GET("/api/status", func(c *gin.Context) {
		c.JSON(200, Map{
			"status": "ok",
		})
	})

	httpService.POST("/api/create/group", func(c *gin.Context) {
		CreateChat(c)
	})

	httpService.GET("/api/group/:id", func(c *gin.Context) {
		groupID, hasGroupID := c.Params.Get("id")

		if !hasGroupID || len(groupID) <= 0 {
			c.JSON(200, Map{
				"message": "group id don't provided",
			})

			return
		}

		querySetOwner := c.Request.URL.Query().Get("setOwner")
		if len(querySetOwner) <= 0 {
			c.JSON(200, Map{
				"message": "setOwner need to be especificated",
			})

			return
		}
	
		setOwner := false
		if strings.Compare(querySetOwner, "true") == 0 {
			setOwner = true
		}

		setOwnerID := c.Request.URL.Query().Get("setOwnerID")
		memberName := c.Request.URL.Query().Get("name")

		ManageWebsocketConnections(groupID, setOwner, setOwnerID, memberName, c.Writer, c.Request)
	})

	go ManagePackets()
}
