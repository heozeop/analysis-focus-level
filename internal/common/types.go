package common

type FocusData struct {
	Date       string         `json:"date"`
	TotalFocus int            `json:"totalFocus"`
	Categories map[string]int `json:"categories"`
	TimeSlots  map[string]int `json:"timeSlots"`
}
