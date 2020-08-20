package theme

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/util"
	"github.com/factly/bindu-server/util/test"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	"gopkg.in/h2non/gock.v1"
)

func TestThemeUpdate(t *testing.T) {
	mock := test.SetupMockDB()
	r := chi.NewRouter()

	r.With(util.CheckUser, util.CheckOrganisation).Mount(url, Router())

	testServer := httptest.NewServer(r)
	gock.New(testServer.URL).EnableNetworking().Persist()
	defer gock.DisableNetworking()
	defer testServer.Close()

	// create httpexpect instance
	e := httpexpect.New(t, testServer.URL)

	var updatedTheme = map[string]interface{}{
		"name": "Dark theme",
		"config": `{"image": { 
			"src": "Images/Sun.png",
			"name": "sun2",
			"hOffset": 25,
			"vOffset": 20,
			"alignment": "left"
		}}`,
	}

	updatedByteData, _ := json.Marshal(updatedTheme["config"])

	t.Run("invalid theme id", func(t *testing.T) {
		e.PUT(urlWithPath).
			WithPath("theme_id", "invalid_id").
			WithHeaders(headers).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("theme record not found", func(t *testing.T) {
		mock.ExpectQuery(selectQuery).
			WithArgs(100, 1).
			WillReturnRows(sqlmock.NewRows(themeProps))

		e.PUT(urlWithPath).
			WithPath("theme_id", "100").
			WithHeaders(headers).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("update theme", func(t *testing.T) {

		mock.ExpectQuery(selectQuery).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "config"}).
				AddRow(1, time.Now(), time.Now(), nil, "Elections", byteData))

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE \"bi_theme\" SET (.+)  WHERE (.+) \"bi_theme\".\"id\" = `).
			WithArgs(updatedByteData, updatedTheme["name"], test.AnyTime{}, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectQuery(selectQuery).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "config"}).
				AddRow(1, time.Now(), time.Now(), nil, updatedTheme["name"], updatedByteData))

		e.PUT(urlWithPath).
			WithPath("theme_id", 1).
			WithHeaders(headers).
			WithJSON(updatedTheme).
			Expect().
			Status(http.StatusOK).JSON().Object().ContainsMap(updatedTheme)

	})

}
