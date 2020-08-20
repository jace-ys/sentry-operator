package sentry

type Avatar struct {
	AvatarType string `json:"avatarType"`
	AvatarUUID string `json:"avatarUuid"`
}

type ListOptions struct {
	Cursor string
}

func Bool(v bool) *bool {
	return &v
}
