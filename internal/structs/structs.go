package structs

type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type FriendsRequest struct {
	SourceId int `json:"source_id"`
	TargetId int `json:"target_id"`
}

type DeleteRequest struct {
	TargetId int `json:"target_id"`
}

type EditAgeRequest struct {
	NewAge int `json:"new_age"`
}
