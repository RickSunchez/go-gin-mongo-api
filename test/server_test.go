package server_test

import (
	"encoding/json"
	"fmt"
	"gin-server/internal/api"
	"gin-server/internal/mongogo"
	"gin-server/internal/structs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	unitTest "github.com/Valiben/gin_unit_test"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type AnswerSuccess struct {
	Ok       bool  `json:"ok"`
	Response gin.H `json:"response"`
}
type AnswerFriends struct {
	Ok       bool           `json:"ok"`
	Response []mongogo.User `json:"response"`
}

var UserList []mongogo.User
var Router *gin.Engine

func init() {
	Router = gin.Default()

	Router.GET("/", api.MethodsList)              // pass
	Router.POST("/create", api.CreateUser)        // pass
	Router.POST("/make_friends", api.MakeFriends) // pass
	Router.DELETE("/user", api.DeleteUser)
	Router.GET("/friends/:user_id", api.GetFriends) // pass
	Router.PUT("/:user_id", api.EditAge)            // pass

	unitTest.SetRouter(Router)
}

func methodListAnswer() string {
	answer := "GET    /                  - methods list\n"
	answer += "POST   /create            - create new user           # {name: <username> string, age: <age> int}\n"
	answer += "POST   /make_friends      - add friend to target user # {source_id: <user_id> int, target_id: <user_id> int}\n"
	answer += "GET    /friends/{user_id} - get friend list\n"
	answer += "DELETE /{user_id}         - delete user               # {target_id: <user_id> int}\n"

	return answer
}
func TestMethodsList(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, methodListAnswer(), w.Body.String())
}

func TestCreateUser(t *testing.T) {
	var userIds []int

	for i := 1; i < 4; i++ {
		resp := structs.CreateUserRequest{
			Name: fmt.Sprintf("Test%d", i),
			Age:  10 * i,
		}

		bytesResp, _ := json.Marshal(resp)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create", strings.NewReader(string(bytesResp)))

		Router.ServeHTTP(w, req)

		var answer AnswerSuccess

		err := json.Unmarshal(w.Body.Bytes(), &answer)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		assert.Equal(t, http.StatusCreated, w.Code)

		userIds = append(userIds, int(answer.Response["user_id"].(float64)))
	}

	mgg, err := mongogo.Init(api.MONGODB)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	assert.Equal(t, mgg.CheckIds(userIds), nil)

	for _, id := range userIds {
		user, err := mgg.GetUser(id)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		UserList = append(UserList, user)
	}

	err = mgg.Disconnect()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestMakeFriends(t *testing.T) {
	mgg, err := mongogo.Init(api.MONGODB)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	for _, sourceUser := range UserList {
		for _, targetUser := range UserList {
			resp := structs.FriendsRequest{
				SourceId: sourceUser.Id,
				TargetId: targetUser.Id,
			}

			bytesResp, _ := json.Marshal(resp)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/make_friends", strings.NewReader(string(bytesResp)))

			Router.ServeHTTP(w, req)

			if sourceUser.Id == targetUser.Id {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				continue
			}

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.NotEqual(t, nil, mgg.FriendExists(sourceUser.Id, targetUser.Id))
		}
	}

	err = mgg.Disconnect()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetFriends(t *testing.T) {
	for _, user := range UserList {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/friends/%d", user.Id), nil)

		Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var answer AnswerFriends

		err := json.Unmarshal(w.Body.Bytes(), &answer)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		assert.Equal(t, 2, len(answer.Response))

		count := 0
		for _, friend := range UserList {
			if user.Id == friend.Id {
				continue
			}

			for _, friendAnswer := range answer.Response {
				if friend.Id == friendAnswer.Id {
					count++
					break
				}
			}
		}

		assert.Equal(t, 2, count)
	}
}

func TestEditAge(t *testing.T) {
	mgg, err := mongogo.Init(api.MONGODB)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	newAge := 70
	for _, user := range UserList {
		resp := structs.EditAgeRequest{
			NewAge: newAge,
		}

		bytesResp, _ := json.Marshal(resp)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d", user.Id), strings.NewReader(string(bytesResp)))

		Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mggUser, err := mgg.GetUser(user.Id)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		assert.Equal(t, newAge, mggUser.Age)
		newAge += 10
	}

	err = mgg.Disconnect()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestDeleteUser(t *testing.T) {
	mgg, err := mongogo.Init(api.MONGODB)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	for _, user := range UserList {
		resp := structs.DeleteRequest{
			TargetId: user.Id,
		}

		bytesResp, _ := json.Marshal(resp)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/user", strings.NewReader(string(bytesResp)))

		Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, nil, mgg.CheckIds([]int{user.Id}))

		for _, friend := range UserList {
			if friend.Id == user.Id {
				continue
			}

			assert.Equal(t, nil, mgg.FriendExists(user.Id, friend.Id))
		}
	}
}
