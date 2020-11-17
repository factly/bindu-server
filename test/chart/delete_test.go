package chart

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/action"
	"github.com/factly/bindu-server/util/test"
	"github.com/gavv/httpexpect/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestChartDelete(t *testing.T) {
	mock := test.SetupMockDB()

	testServer := httptest.NewServer(action.RegisterRoutes())
	gock.New(testServer.URL).EnableNetworking().Persist()
	defer gock.DisableNetworking()
	defer testServer.Close()

	// create httpexpect instance
	e := httpexpect.New(t, testServer.URL)

	t.Run("invalid chart id", func(t *testing.T) {

		e.DELETE(path).
			WithPath("chart_id", "invalid_id").
			WithHeaders(headers).
			Expect().
			Status(http.StatusNotFound)

	})

	t.Run("chart record not found", func(t *testing.T) {

		recordNotFoundMock(mock)

		e.DELETE(path).
			WithPath("chart_id", "100").
			WithHeaders(headers).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("chart record deleted", func(t *testing.T) {
		chartSelectMock(mock)

		mock.ExpectBegin()
		mock.ExpectExec(deleteQuery).
			WithArgs(test.AnyTime{}, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		e.DELETE(path).
			WithPath("chart_id", 1).
			WithHeaders(headers).
			Expect().
			Status(http.StatusOK)
	})

}
