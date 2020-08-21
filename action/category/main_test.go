package category

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/bindu-server/util/test"
	"github.com/joho/godotenv"
	"gopkg.in/h2non/gock.v1"
)

var headers = map[string]string{
	"X-Organisation": "1",
	"X-User":         "1",
}

var data = map[string]interface{}{
	"name": "Politics",
	"slug": "politics",
}

var categoryWithoutSlug = map[string]interface{}{
	"name": "Politics",
	"slug": "",
}

var categoryProps = []string{"id", "created_at", "updated_at", "deleted_at", "name", "slug"}

var selectQuery = regexp.QuoteMeta(`SELECT * FROM "bi_category"`)
var chartQuery = regexp.QuoteMeta(`SELECT count(*) FROM "bi_chart" INNER JOIN "bi_chart_category"`)
var deleteQuery = regexp.QuoteMeta(`UPDATE "bi_category" SET "deleted_at"=`)
var countQuery = regexp.QuoteMeta(`SELECT count(*) FROM "bi_category"`)
var paginationQuery = `SELECT \* FROM "bi_category" (.+) LIMIT 1 OFFSET 1`

var url = "/categories"
var urlWithPath = "/categories/{category_id}"

func categorySelectMock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(selectQuery).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "slug"}).
			AddRow(1, time.Now(), time.Now(), nil, data["name"], data["slug"]))
}

func categoryChartExpect(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(chartQuery).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func TestMain(m *testing.M) {

	godotenv.Load("../../.env")

	// Mock kavach server and allowing persisted external traffic
	defer gock.Disable()
	test.MockServer()
	defer gock.DisableNetworking()

	exitValue := m.Run()

	os.Exit(exitValue)
}
