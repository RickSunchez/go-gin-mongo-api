package api

import (
	"encoding/json"
	"fmt"
	"gin-server/internal/errors"
	"gin-server/internal/mongogo"
	"gin-server/internal/structs"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	MONGODB string = "localhost:27017"
)

var HTTPerr errors.HTTPErrors

func MethodsList(c *gin.Context) {
	answer := "GET    /                  - methods list\n"
	answer += "POST   /create            - create new user           # {name: <username> string, age: <age> int}\n"
	answer += "POST   /make_friends      - add friend to target user # {source_id: <user_id> int, target_id: <user_id> int}\n"
	answer += "GET    /friends/{user_id} - get friend list\n"
	answer += "DELETE /{user_id}         - delete user               # {target_id: <user_id> int}\n"

	c.String(http.StatusOK, answer)
}

func CreateUser(c *gin.Context) {
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	var user mongogo.User
	err = json.Unmarshal(rawData, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	mgg, err := mongogo.Init(MONGODB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	userId, err := mgg.NewUser(user.Name, user.Age)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	if mgg.Disconnect() != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok": true,
		"response": gin.H{
			"message": fmt.Sprintf("User %s created with id: %d", user.Name, userId),
			"user_id": userId,
		},
	})
}

func MakeFriends(c *gin.Context) {
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	var request structs.FriendsRequest
	err = json.Unmarshal(rawData, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	mgg, err := mongogo.Init(MONGODB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	err = mgg.AddFriend(request.SourceId, request.TargetId)
	if _, ok := err.(*errors.UndefinedIndexes); ok {
		c.JSON(http.StatusBadRequest, HTTPerr.ErrorJSON(err))
		return
	} else if _, ok := err.(*errors.FriendsExists); ok {
		c.JSON(http.StatusBadRequest, HTTPerr.ErrorJSON(err))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	if mgg.Disconnect() != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok":       true,
		"response": fmt.Sprintf("User %d added as friend to %d", request.SourceId, request.TargetId),
	})
}

func DeleteUser(c *gin.Context) {
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	var request structs.DeleteRequest
	err = json.Unmarshal(rawData, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	mgg, err := mongogo.Init(MONGODB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	userName, err := mgg.DelUser(request.TargetId)
	if _, ok := err.(*errors.UndefinedIndexes); ok {
		c.String(http.StatusBadRequest, "Error: %v", err)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	if mgg.Disconnect() != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"response": fmt.Sprintf("User %s deleted", userName),
	})
}

func GetFriends(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, HTTPerr.ErrorJSON(err))
		return
	}

	mgg, err := mongogo.Init(MONGODB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	friends, err := mgg.GetFriends(userId)
	if _, ok := err.(*errors.UndefinedIndexes); ok {
		c.String(http.StatusBadRequest, "Error: %v", err)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	if mgg.Disconnect() != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"response": friends,
	})
}

func EditAge(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, HTTPerr.ErrorJSON(err))
		return
	}

	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	var request structs.EditAgeRequest
	err = json.Unmarshal(rawData, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	mgg, err := mongogo.Init(MONGODB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	err = mgg.UpdateAge(userId, request.NewAge)
	if _, ok := err.(*errors.UndefinedIndexes); ok {
		c.String(http.StatusBadRequest, "Error: %v", err)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	if mgg.Disconnect() != nil {
		c.JSON(http.StatusInternalServerError, HTTPerr.ErrorJSON(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"response": fmt.Sprintf("Age updated for user %d", userId),
	})
}
