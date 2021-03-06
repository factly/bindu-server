package category

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/action"
	"github.com/factly/bindu-server/util/test"
	"github.com/gavv/httpexpect/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestCategoryUpdate(t *testing.T) {
	mock := test.SetupMockDB()
	test.MockServers()
	testServer := httptest.NewServer(action.RegisterRoutes())
	gock.New(testServer.URL).EnableNetworking().Persist()
	defer gock.DisableNetworking()
	defer testServer.Close()

	// create httpexpect instance
	e := httpexpect.New(t, testServer.URL)

	t.Run("invalid category id", func(t *testing.T) {
		test.CheckSpace(mock)
		e.PUT(path).
			WithPath("category_id", "invalid_id").
			WithHeaders(headers).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("cannot decode category", func(t *testing.T) {
		test.CheckSpace(mock)

		e.PUT(path).
			WithPath("category_id", 1).
			WithHeaders(headers).
			Expect().
			Status(http.StatusUnprocessableEntity)

	})

	t.Run("Unprocessable category", func(t *testing.T) {
		test.CheckSpace(mock)

		e.PUT(path).
			WithPath("category_id", 1).
			WithHeaders(headers).
			WithJSON(invalidData).
			Expect().
			Status(http.StatusUnprocessableEntity)

	})

	t.Run("category record not found", func(t *testing.T) {
		test.CheckSpace(mock)
		recordNotFoundMock(mock)

		e.PUT(path).
			WithPath("category_id", "100").
			WithJSON(data).
			WithHeaders(headers).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("update category", func(t *testing.T) {

		test.CheckSpace(mock)
		updatedCategory := map[string]interface{}{
			"name":            "Politics",
			"slug":            "politics",
			"is_for_template": true,
		}

		SelectMock(mock)

		categoryUpdateMock(mock, updatedCategory)

		selectAfterUpdate(mock, updatedCategory)
		mock.ExpectCommit()

		e.PUT(path).
			WithPath("category_id", 1).
			WithHeaders(headers).
			WithJSON(updatedCategory).
			Expect().
			Status(http.StatusOK).JSON().Object().ContainsMap(updatedCategory)

	})

	t.Run("update category by id with empty slug", func(t *testing.T) {

		test.CheckSpace(mock)
		updatedCategory := map[string]interface{}{
			"name":            "Politics",
			"slug":            "politics-1",
			"is_for_template": true,
		}
		SelectMock(mock)

		mock.ExpectQuery(`SELECT slug, space_id FROM "bi_category"`).
			WithArgs("politics%", 1).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(1, time.Now(), time.Now(), nil, 1, 1, "Politics", "politics", true, 1))

		categoryUpdateMock(mock, updatedCategory)

		selectAfterUpdate(mock, updatedCategory)
		mock.ExpectCommit()

		e.PUT(path).
			WithPath("category_id", 1).
			WithHeaders(headers).
			WithJSON(dataWithoutSlug).
			Expect().
			Status(http.StatusOK).JSON().Object().ContainsMap(updatedCategory)

	})

	t.Run("update category with different slug", func(t *testing.T) {
		test.CheckSpace(mock)
		updatedCategory := map[string]interface{}{
			"name":            "Politics",
			"slug":            "testing-slug",
			"is_for_template": true,
		}
		SelectMock(mock)

		mock.ExpectQuery(`SELECT slug, space_id FROM "bi_category"`).
			WithArgs(fmt.Sprint(updatedCategory["slug"], "%"), 1).
			WillReturnRows(sqlmock.NewRows([]string{"slug", "space_id"}))

		categoryUpdateMock(mock, updatedCategory)

		selectAfterUpdate(mock, updatedCategory)
		mock.ExpectCommit()

		e.PUT(path).
			WithPath("category_id", 1).
			WithHeaders(headers).
			WithJSON(updatedCategory).
			Expect().
			Status(http.StatusOK).JSON().Object().ContainsMap(updatedCategory)

	})

}
