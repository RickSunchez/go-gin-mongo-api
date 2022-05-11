package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UndefinedIndexes struct {
	Indexes []int
}

func (ui *UndefinedIndexes) Error() string {
	return fmt.Sprintf("Undefined indexes: %v", ui.Indexes)
}

type InternarMongoError struct {
	Err error
}

func (ime *InternarMongoError) Error() string {
	return fmt.Sprintf("Iternal mongodb error: %v", ime.Err)
}

type FriendsExists struct {
	SourceId int
	TargetId int
}

func (fe *FriendsExists) Error() string {
	return fmt.Sprintf("User %d is already in friend list of %d", fe.SourceId, fe.TargetId)
}

type HTTPErrors struct{}

func (he *HTTPErrors) InternalError(w http.ResponseWriter, message interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%v", message)))
}
func (he *HTTPErrors) ErrorJSON(err error) gin.H {
	return gin.H{
		"ok":       false,
		"response": err.Error(),
	}
}
