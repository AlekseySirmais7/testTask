package models

type Statistic struct {
	AddCount    int `json:"add_count"`
	UpdateCount int `json:"update_count"`
	DeleteCount int `json:"delete_count"`
}
