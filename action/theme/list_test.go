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

func TestThemeList(t *testing.T) {
	mock := test.SetupMockDB()
	r := chi.NewRouter()

	r.With(util.CheckUser, util.CheckOrganisation).Mount(url, Router())

	testServer := httptest.NewServer(r)
	gock.New(testServer.URL).EnableNetworking().Persist()
	defer gock.DisableNetworking()
	defer testServer.Close()

	// create httpexpect instance
	e := httpexpect.New(t, testServer.URL)

	themelist := []map[string]interface{}{
		{"name": "Test Theme 1", "config": `{"image": { 
			"src": "Images/Sun.png",
			"name": "sun1",
			"hOffset": 250,
			"vOffset": 250,
			"alignment": "center"
		}}`},
		{"name": "Test Theme 2", "config": `{"image": { 
			"src": "Images/Sun.png",
			"name": "sun2",
			"hOffset": 250,
			"vOffset": 250,
			"alignment": "center"
		}}`},
	}

	byteData0, _ := json.Marshal(themelist[0]["config"])
	byteData1, _ := json.Marshal(themelist[1]["config"])

	t.Run("get empty list of themes", func(t *testing.T) {

		mock.ExpectQuery(countQuery).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow("0"))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows(themeProps))

		e.GET(url).
			WithHeaders(headers).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			ContainsMap(map[string]interface{}{"total": 0})

		mock.ExpectationsWereMet()
	})

	t.Run("get non-empty list of themes", func(t *testing.T) {

		mock.ExpectQuery(countQuery).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(themelist)))

		mock.ExpectQuery(selectQuery).
			WillReturnRows(sqlmock.NewRows(themeProps).
				AddRow(1, time.Now(), time.Now(), nil, themelist[0]["name"], byteData0).
				AddRow(2, time.Now(), time.Now(), nil, themelist[1]["name"], byteData1))

		e.GET(url).
			WithHeaders(headers).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			ContainsMap(map[string]interface{}{"total": len(themelist)}).
			Value("nodes").
			Array().
			Element(0).
			Object().
			ContainsMap(themelist[0])

		mock.ExpectationsWereMet()
	})

	t.Run("get themes with pagination", func(t *testing.T) {
		mock.ExpectQuery(countQuery).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(themelist)))

		mock.ExpectQuery(paginationQuery).
			WillReturnRows(sqlmock.NewRows(themeProps).
				AddRow(2, time.Now(), time.Now(), nil, themelist[1]["name"], byteData1))

		e.GET(url).
			WithQueryObject(map[string]interface{}{
				"limit": "1",
				"page":  "2",
			}).
			WithHeaders(headers).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			ContainsMap(map[string]interface{}{"total": len(themelist)}).
			Value("nodes").
			Array().
			Element(0).
			Object().
			ContainsMap(themelist[1])

		mock.ExpectationsWereMet()

	})
}
