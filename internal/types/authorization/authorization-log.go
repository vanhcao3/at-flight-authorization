package types

type AuthorizationLog struct {
	Result    string `json:"result,omitempty"`
	UserId    string `json:"userid,omitempty"`
	Username  string `json:"username,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Action    string `json:"action,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Client    string `json:"client,omitempty"`
	URI       string `json:"uri,omitempty"`
}
