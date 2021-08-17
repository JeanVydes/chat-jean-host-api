package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Group struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	OwnerID string           `json:"owner_id"`
	Members map[string]*User `json:"members"`
}

func CreateChat(c *gin.Context) {
	groupName := c.Request.URL.Query().Get("groupName")

	if len(groupName) <= 0 {
		groupName = "A Cool Group!"
	}
	
	if len(groupName) > 30 {
		groupName = groupName[0:30]
	}

	ownerID := fmt.Sprintf("%v", RandomID())

	group := &Group{
		ID:      fmt.Sprintf("%v", RandomID()),
		Name:    groupName,
		OwnerID: ownerID,
		Members: make(map[string]*User),
	}

	Groups[group.ID] = group

	c.JSON(200, Map{
		"code":    0,
		"message": "Group created succesfully!",
		"data": Map{
			"group": group,
		},
	})
}
