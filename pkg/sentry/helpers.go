package sentry

type Avatar struct {
	AvatarType string `json:"avatarType"`
	AvatarUUID string `json:"avatarUuid"`
}

func Bool(v bool) *bool {
	return &v
}
