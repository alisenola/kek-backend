package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kek-backend/internal/account"
	accountDBMock "kek-backend/internal/account/database/mocks"
	accountModel "kek-backend/internal/account/model"
	"kek-backend/internal/alert/database"
	alertDBMock "kek-backend/internal/alert/database/mocks"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/config"
	"kek-backend/pkg/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"go.uber.org/zap/zapcore"
)

var (
	dUser = accountModel.Account{
		ID:       1,
		Username: "user1",
		Email:    "user1@gmail.com",
		Password: "$2a$10$lsYsLv8nGPM0.R.ft4sgpe3OP7..KL3ZJqqhSVCKTEnSCMUztoUcW",
		Bio:      "I am working!",
	}
	dUserRawPass = "user1"

	dAlert = model.Alert{
		ID:        1,
		Slug:      "how-to-train-your-dragon",
		Title:     "How to train your dragon",
		Body:      "You have to believe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
)

type HandlerSuite struct {
	suite.Suite
	r         *gin.Engine
	handler   *Handler
	db        *alertDBMock.AlertDB
	accountDB *accountDBMock.AccountDB
}

func (s *HandlerSuite) SetupSuite() {
	logging.SetLevel(zapcore.FatalLevel)
}

func (s *HandlerSuite) SetupTest() {
	cfg, err := config.Load("")
	s.NoError(err)

	s.db = &alertDBMock.AlertDB{}
	s.handler = NewHandler(s.db)
	s.accountDB = &accountDBMock.AccountDB{}
	s.accountDB.On("FindByEmail", mock.Anything, mock.MatchedBy(func(email string) bool {
		return email == dUser.Email
	})).Return(&dUser, nil)

	jwtMiddleware, err := account.NewAuthMiddleware(cfg, s.accountDB)
	s.NoError(err)

	gin.SetMode(gin.TestMode)
	s.r = gin.Default()

	RouteV1(cfg, s.handler, s.r, jwtMiddleware)

	accountHandler := account.NewHandler(s.accountDB)
	account.RouteV1(cfg, accountHandler, s.r, jwtMiddleware)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

// TODO : test failures

func (s *HandlerSuite) TestSaveAlert() {
	// given
	s.db.On("SaveAlert", mock.Anything, mock.Anything).Return(nil)

	// when
	requestBody := map[string]interface{}{
		"alert": map[string]interface{}{
			"title": dAlert.Title,
			"body":  dAlert.Body,
		},
	}
	b, _ := json.Marshal(&requestBody)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/api/alerts", bytes.NewBuffer(b))
	req.Header.Add("Authorization", "Bearer "+s.getBearerToken())

	s.r.ServeHTTP(res, req)

	// then
	// 1) method called
	alertMatcher := alertMatcher(dAlert.Title, dAlert.Body, &dUser)
	s.db.AssertCalled(s.T(), "SaveAlert", mock.Anything, mock.MatchedBy(alertMatcher))
	// 2) status code
	s.Equal(http.StatusCreated, res.Code)
	// 3) response
	jsonVal := res.Body.String()
	s.assertAlertResponse(&dAlert, gjson.Parse(jsonVal).Get("alert"))
}

func (s *HandlerSuite) TestAlertBySlug() {
	// given
	s.db.On("FindAlertBySlug", mock.Anything, dAlert.Slug).Return(&dAlert, nil)

	// when
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/api/alerts/"+dAlert.Slug, nil)

	s.r.ServeHTTP(res, req)

	// then
	// 1) method called
	s.db.AssertCalled(s.T(), "FindAlertBySlug", mock.Anything, dAlert.Slug)
	// 2) status code
	s.Equal(http.StatusOK, res.Code)
	// 3) body
	jsonVal := res.Body.String()
	s.assertAlertResponse(&dAlert, gjson.Parse(jsonVal).Get("alert"))
}

func (s *HandlerSuite) TestAlerts() {
	criteria := database.IterateAlertCriteria{
		Account: dAlert.Account.Username,
		Offset:  0,
		Limit:   5,
	}
	s.db.On("FindAlerts", mock.Anything, criteria).Return([]*model.Alert{&dAlert}, int64(1), nil)

	// when
	url := fmt.Sprintf("/v1/api/alerts?tag=%s&account=%s&offset=%d&limit=%d",
		criteria.Account, criteria.Offset, criteria.Limit)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)

	s.r.ServeHTTP(res, req)

	// then
	// 1) method called
	s.db.AssertCalled(s.T(), "FindAlerts", mock.Anything, criteria)
	// 2) status code
	s.Equal(http.StatusOK, res.Code)
	// 3) body
	jsonVal := res.Body.String()
	result := gjson.Parse(jsonVal)
	s.Equal(int64(1), result.Get("alertsCount").Int())

	alertsResult := result.Get("alerts").Array()
	s.Equal(1, len(alertsResult))
	s.assertAlertResponse(&dAlert, alertsResult[0])
}

func (s *HandlerSuite) TestDeleteAlert() {
	// given
	s.db.On("RunInTx", mock.Anything, mock.Anything).Return(nil)

	// when
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/api/alerts/"+dAlert.Slug, nil)
	req.Header.Add("Authorization", "Bearer "+s.getBearerToken())

	s.r.ServeHTTP(res, req)

	// then
	// 1) method called
	s.db.AssertCalled(s.T(), "RunInTx", mock.Anything, mock.Anything)
	// 2) status code
	s.Equal(http.StatusOK, res.Code)
	// 3) body
	s.Empty(res.Body.Bytes())
}

func (s *HandlerSuite) assertAlertResponse(alert *model.Alert, result gjson.Result) {
	s.Equal(slug.Make(alert.Title), result.Get("slug").String())
	s.Equal(alert.Title, result.Get("title").String())
	s.Equal(alert.Body, result.Get("body").String())

	s.True(result.Get("createdAt").Exists())
	s.True(result.Get("updatedAt").Exists())
	s.Equal(alert.Account.Username, result.Get("account.username").String())
	s.Equal(alert.Account.Bio, result.Get("account.bio").String())
	s.Equal(alert.Account.Image, result.Get("account.image").String())
}

func (s *HandlerSuite) getBearerToken() string {
	body := map[string]interface{}{
		"user": map[string]interface{}{
			"email":    dUser.Email,
			"password": dUserRawPass,
		},
	}
	b, _ := json.Marshal(body)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/api/users/login", bytes.NewBuffer(b))
	s.r.ServeHTTP(res, req)

	s.Equal(http.StatusOK, res.Code)
	return gjson.Get(res.Body.String(), "token").String()
}

func alertMatcher(title, body string, account *accountModel.Account) func(a *model.Alert) bool {
	return func(a *model.Alert) bool {
		if a.Slug != slug.Make(title) || a.Title != title || a.Body != body {
			return false
		}
		if a.Account.Username != account.Username || a.Account.Bio != account.Bio || a.Account.Image != account.Image {
			return false
		}
		return true
	}
}
