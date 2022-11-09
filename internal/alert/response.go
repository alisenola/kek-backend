package alert

import (
	accountModel "kek-backend/internal/account/model"
	"kek-backend/internal/alert/model"
	"time"
)

type AlertResponse struct {
	Alert Alert `json:"alert"`
}

type AlertsResponse struct {
	Alert       []Alert `json:"alerts"`
	AlertsCount int64   `json:"alertsCount"`
}

type Alert struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	PairAddress    string    `json:"pairAddress"`
	AlertType      string    `json:"alertType"`
	AlertValue     string    `json:"alertValue"`
	AlertOption    string    `json:"alertOption"`
	ExpirationTime time.Time `json:"expirationTime"`
	AlertActions   string    `json:"alertActions"`
	AlertStatus    string    `json:"alertStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Account        accountModel.Account
}

// NewAlertsResponse converts alert models and total count to AlertsResponse
func NewAlertsResponse(alerts []*model.Alert, total int64) *AlertsResponse {
	var a []Alert
	for _, alert := range alerts {
		a = append(a, NewAlertResponse(alert).Alert)
	}

	return &AlertsResponse{
		Alert:       a,
		AlertsCount: total,
	}
}

// NewAlertResponse converts alert model to AlertResponse
func NewAlertResponse(a *model.Alert) *AlertResponse {
	return &AlertResponse{
		Alert: Alert{
			Slug:           a.Slug,
			Title:          a.Title,
			Body:           a.Body,
			PairAddress:    a.PairAddress,
			AlertType:      a.AlertType,
			AlertValue:     a.AlertValue,
			AlertOption:    a.AlertOption,
			ExpirationTime: a.ExpirationTime,
			AlertActions:   a.AlertActions,
			AlertStatus:    a.AlertStatus,
			CreatedAt:      a.CreatedAt,
			UpdatedAt:      a.UpdatedAt,
			Account:        a.Account,
		},
	}
}
