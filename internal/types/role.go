package types

type Role struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Attributes  interface{} `json:"attributes"`
}

type RoleMetadata struct {
	Id          string `json:"id"`
	ContainerId string `json:"containerId"`
	Name        string `json:"name"`
}

type EffectiveRoleMetadata struct {
	Id   string `json:"id"`
	Role string `json:"role"`
}
