package dto

type PanelGroupDateOutput struct {
	ServiceNumber int64 `json:"service_number"`
	AppNumber int64 `json:"app_number"`
	CurrentQPS int64 `json:"current_qps"`
	TodayRequestNumber int64 `json:"today_request_number"`
}
type DashServiceStatisticsItemOutput struct {
	Name    string `json:"name"`
	LoadType int `json:"load_type"`
	Value   int64 `json:"value"`
}
type DashServiceStatisticsOutput struct {
	Legend    []string  `json:"legend"`
	Data      []DashServiceStatisticsItemOutput `json:"data"`
	CurrentQPS int64 `json:"current_qps"`
	TodayRequestNumber int64 `json:"today_request_number"`
}