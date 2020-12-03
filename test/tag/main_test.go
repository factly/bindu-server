package tag

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/util/test"
	"gopkg.in/h2non/gock.v1"
)

var headers = map[string]string{
	"X-Organisation": "1",
	"X-User":         "1",
}

var data = map[string]interface{}{
	"name": "Elections",
	"slug": "elections",
}

var dataWithoutSlug = map[string]interface{}{
	"name": "Elections",
	"slug": "",
}

var invalidData = map[string]interface{}{
	"name": "ab",
}

var columns = []string{"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id", "name", "slug"}

var selectQuery = regexp.QuoteMeta(`SELECT * FROM "bi_tag"`)
var deleteQuery = regexp.QuoteMeta(`UPDATE "bi_tag" SET "deleted_at"=`)
var paginationQuery = `SELECT \* FROM "bi_tag" (.+) LIMIT 1 OFFSET 1`

var basePath = "/tags"
var path = "/tags/{tag_id}"

func slugCheckMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT slug, organisation_id FROM "bi_tag"`)).
		WithArgs(fmt.Sprint(data["slug"], "%"), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "organisation_id", "name", "slug"}))
}

func tagInsertMock(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "bi_tag"`).
		WithArgs(test.AnyTime{}, test.AnyTime{}, nil, 1, 1, data["name"], data["slug"], "", 1).
		WillReturnRows(sqlmock.
			NewRows([]string{"id"}).
			AddRow(1))
	mock.ExpectCommit()
}

//check tag exits or not
func recordNotFoundMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(selectQuery).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns))
}

func tagSelectMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(selectQuery).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, time.Now(), time.Now(), nil, 1, 1, data["name"], data["slug"]))
}

func tagSelectWithOutOrg(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(selectQuery).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, time.Now(), time.Now(), nil, 1, 1, data["name"], data["slug"]))
}

// check tag associated with any chart before deleting
func tagChartExpect(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "bi_chart" JOIN "bi_chart_tag"`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func tagUpdateMock(mock sqlmock.Sqlmock, tag map[string]interface{}) {
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE \"bi_tag\"`).
		WithArgs(test.AnyTime{}, 1, tag["name"], tag["slug"], 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func tagCountQuery(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "bi_tag"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func selectAfterUpdate(mock sqlmock.Sqlmock, tag map[string]interface{}) {
	mock.ExpectQuery(selectQuery).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, time.Now(), time.Now(), nil, 1, 1, tag["name"], tag["slug"]))
}

func TestMain(m *testing.M) {

	test.SetEnv()

	// Mock kavach server and allowing persisted external traffic
	defer gock.Disable()
	test.MockServer()
	defer gock.DisableNetworking()

	exitValue := m.Run()

	os.Exit(exitValue)
}
