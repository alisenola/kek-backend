package alert

import (
	"context"
	"kek-backend/internal/account"
	alertDB "kek-backend/internal/alert/database"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/config"
	"kek-backend/internal/database"
	"kek-backend/internal/middleware"
	"kek-backend/internal/middleware/handler"
	"kek-backend/pkg/logging"
	"kek-backend/pkg/validate"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
)

type Handler struct {
	alertDB alertDB.AlertDB
}

// saveAlert handles POST /v1/api/alerts
func (h *Handler) saveAlert(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		// bind
		type RequestBody struct {
			Alert struct {
				Title          string    `json:"title" binding:"required,min=5"`
				Body           string    `json:"body" binding:"required"`
				PairAddress    string    `json:"pairAddress" binding:"required,min=20"`
				AlertType      string    `json:"alertType" binding:"required,min=3"`
				AlertValue     string    `json:"alertValue" binding:"required"`
				AlertOption    string    `json:"alertOption" binding:"required"`
				ExpirationTime time.Time `json:"expirationTime" binding:"required"`
				AlertActions   string    `json:"alertActions" binding:"required"`
			} `json:"alert"`
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("alert.handler.register failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.Alert, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid alert request in body", details)
		}

		// save alert
		currentUser := account.MustCurrentUser(c)
		alert := model.Alert{
			Slug:           slug.Make(body.Alert.Title),
			Title:          body.Alert.Title,
			Body:           body.Alert.Body,
			PairAddress:    body.Alert.PairAddress,
			AlertType:      body.Alert.AlertType,
			AlertValue:     body.Alert.AlertValue,
			AlertOption:    body.Alert.AlertOption,
			ExpirationTime: body.Alert.ExpirationTime,
			AlertActions:   body.Alert.AlertActions,
			AlertStatus:    "active",
			AccountId:      currentUser.ID,
		}
		err := h.alertDB.SaveAlert(c.Request.Context(), &alert)
		if err != nil {
			if database.IsKeyConflictErr(err) {
				return handler.NewErrorResponse(http.StatusConflict, handler.DuplicateEntry, "duplicate alert title", nil)
			}
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusCreated, NewAlertResponse(&alert))
	})
}

// alertBySlug handles GET /v1/api/alerts/:slug
func (h *Handler) alertBySlug(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		// bind
		type RequestUri struct {
			Slug string `uri:"slug"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("alert.handler.alertBySlug failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid alert request in uri", details)
		}

		// find
		alert, err := h.alertDB.FindAlertBySlug(c.Request.Context(), uri.Slug)
		if err != nil {
			if database.IsRecordNotFoundErr(err) {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "not found alert", nil)
			}
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, NewAlertResponse(alert))
	})
}

// alerts handles GET /v1/api/alerts
func (h *Handler) alerts(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type QueryParameter struct {
			Tag     []string `form:"tag" binding:"omitempty,dive,max=10"`
			Account string   `form:"account" binding:"omitempty"`
			Limit   string   `form:"limit,default=5" binding:"numeric"`
			Offset  string   `form:"offset,default=0" binding:"numeric"`
		}
		var query QueryParameter
		if err := c.ShouldBindQuery(&query); err != nil {
			logger.Errorw("alert.handler.alerts failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&query, "form", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid alert request in query", details)
		}

		account, err := strconv.ParseUint(query.Account, 10, 64)
		limit, err := strconv.ParseUint(query.Limit, 10, 64)
		if err != nil {
			limit = 5
		}
		offset, err := strconv.ParseUint(query.Offset, 10, 64)
		if err != nil {
			offset = 0
		}
		criteria := alertDB.IterateAlertCriteria{
			Account: uint(account),
			Offset:  uint(offset),
			Limit:   uint(limit),
		}
		alerts, total, err := h.alertDB.FindAlerts(c.Request.Context(), criteria)
		if err != nil {
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, NewAlertsResponse(alerts, total))
	})
}

// deleteAlert handles DELETE /v1/api/alerts/:slug
func (h *Handler) deleteAlert(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		// bind
		type RequestUri struct {
			Slug string `uri:"slug" binding:"required"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("alert.handler.deleteAlert failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid alert request in uri", details)
		}

		// delete alert in transaction
		currentUser := account.MustCurrentUser(c)
		err := h.alertDB.RunInTx(c.Request.Context(), func(ctx context.Context) error {
			// delete a alert
			if err := h.alertDB.DeleteAlertBySlug(ctx, currentUser.ID, uri.Slug); err != nil {
				return err
			}

			logger.Debugw("alert.handler.deleteAlert success to delete a alert", "comments")
			return nil
		})
		if err != nil {
			logger.Errorw("alert.handler.deleteAlert failed to delete a alert", "err", err)
			if database.IsRecordNotFoundErr(errors.Cause(err)) {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "not found alert", nil)
			}
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, nil)
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	v1 := r.Group("v1/api")
	timeout := time.Duration(cfg.ServerConfig.WriteTimeoutSecs) * time.Second
	v1.Use(middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(timeout))

	alertV1 := v1.Group("alerts")
	// anonymous
	alertV1.Use()
	{
		alertV1.GET(":slug", h.alertBySlug)
		alertV1.GET("", h.alerts)
	}

	// auth required
	alertV1.Use(auth.MiddlewareFunc())
	{
		alertV1.POST("", h.saveAlert)
		alertV1.DELETE(":slug", h.deleteAlert)
	}
}

func NewHandler(alertDB alertDB.AlertDB) *Handler {
	StartCron(alertDB)
	return &Handler{
		alertDB: alertDB,
	}
}
