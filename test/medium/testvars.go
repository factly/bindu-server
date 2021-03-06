package medium

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/util/test"
)

var headers = map[string]string{
	"X-Space": "1",
	"X-User":  "1",
}
var invalidData = map[string]interface{}{
	"nam": "po",
}

var data = map[string]interface{}{
	"name":        "Image",
	"slug":        "image",
	"type":        "jpg",
	"title":       "Sample image",
	"description": "desc",
	"caption":     "sample",
	"alt_text":    "sample",
	"file_size":   100,
	"url": postgres.Jsonb{
		RawMessage: []byte(`{"raw":"http://testimage.com/test.jpg"}`),
	},
	"dimensions": "testdims",
}

var mediumWithoutSlug = map[string]interface{}{
	"name":        "Image",
	"type":        "jpg",
	"title":       "Sample image",
	"description": "desc",
	"caption":     "sample",
	"alt_text":    "sample",
	"file_size":   100,
	"url": postgres.Jsonb{
		RawMessage: []byte(`{"raw":"http://testimage.com/test.jpg"}`),
	},
	"dimensions": "testdims",
}

var columns = []string{"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id", "name", "slug", "type", "title", "description", "caption", "alt_text", "file_size", "url", "dimensions", "space_id"}

var selectQuery = regexp.QuoteMeta(`SELECT * FROM "bi_medium"`)
var chartQuery = regexp.QuoteMeta(`SELECT count(1) FROM "bi_chart"`)
var deleteQuery = regexp.QuoteMeta(`UPDATE "bi_medium" SET "deleted_at"=`)
var paginationQuery = `SELECT \* FROM "bi_medium" (.+) LIMIT 1 OFFSET 1`

var basePath = "/media"
var path = "/media/{medium_id}"

func SelectMock(mock sqlmock.Sqlmock, args ...driver.Value) {
	mock.ExpectQuery(selectQuery).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, time.Now(), time.Now(), nil, 1, 1, data["name"], data["slug"], data["type"], data["title"], data["description"], data["caption"], data["alt_text"], data["file_size"], data["url"], data["dimensions"], 1))
}

func mediumChartExpect(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(chartQuery).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

//check medium exits or not
func recordNotFoundMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(selectQuery).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns))
}

func slugCheckMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT slug, space_id FROM "bi_medium"`)).
		WithArgs(fmt.Sprint(data["slug"], "%"), 1).
		WillReturnRows(sqlmock.NewRows(columns))
}

func mediumInsertMock(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "bi_medium"`).
		WithArgs(test.AnyTime{}, test.AnyTime{}, nil, 1, 1, data["name"], data["slug"], data["type"], data["title"], data["description"], data["caption"], data["alt_text"], data["file_size"], data["url"], data["dimensions"], 1).
		WillReturnRows(sqlmock.
			NewRows([]string{"id"}).
			AddRow(1))
	mock.ExpectCommit()
}

func mediumCountQuery(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "bi_medium"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func mediumUpdateMock(mock sqlmock.Sqlmock, medium map[string]interface{}) {
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE \"bi_medium\"`).
		WithArgs(test.AnyTime{}, 1, medium["name"], medium["slug"], medium["type"], medium["title"], medium["description"], medium["caption"], medium["alt_text"], medium["file_size"], medium["url"], medium["dimensions"], 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func selectAfterUpdate(mock sqlmock.Sqlmock, medium map[string]interface{}) {
	mock.ExpectQuery(selectQuery).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, time.Now(), time.Now(), nil, 1, 1, medium["name"], medium["slug"], medium["type"], medium["title"], medium["description"], medium["caption"], medium["alt_text"], medium["file_size"], medium["url"], medium["dimensions"], 1))
}
