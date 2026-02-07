package models

type TabState struct {
	TabID     string `json:"tabId"`
	Name      string `json:"name"`
	Component string `json:"component"`
	Props     string `json:"props"`
	State     string `json:"state"`
	IsActive  bool   `json:"isActive"`
	Order     int    `json:"order"`
}
