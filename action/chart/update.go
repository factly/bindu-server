package chart

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/factly/bindu-server/config"
	"github.com/factly/bindu-server/model"
	"github.com/factly/bindu-server/util"
	"github.com/factly/bindu-server/util/slug"
	"github.com/factly/x/errorx"
	"github.com/factly/x/loggerx"
	"github.com/factly/x/renderx"
	"github.com/factly/x/validationx"
	"github.com/go-chi/chi"
	"gorm.io/gorm"
)

// update - Update chart by id
// @Summary Update a chart by id
// @Description Update chart by ID
// @Tags Chart
// @ID update-chart-by-id
// @Produce json
// @Consume json
// @Param X-User header string true "User ID"
// @Param chart_id path string true "Chart ID"
// @Param X-Organisation header string true "Organisation ID"
// @Param Chart body chart false "Chart"
// @Success 200 {object} model.Chart
// @Router /charts/{chart_id} [put]
func update(w http.ResponseWriter, r *http.Request) {
	chartID := chi.URLParam(r, "chart_id")
	id, err := strconv.Atoi(chartID)

	if err != nil {
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.InvalidID()))
		return
	}

	oID, err := util.GetOrganisation(r.Context())
	if err != nil {
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.Unauthorized()))
		return
	}

	uID, err := util.GetUser(r.Context())
	if err != nil {
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.Unauthorized()))
		return
	}

	chart := &chart{}
	err = json.NewDecoder(r.Body).Decode(&chart)
	if err != nil {
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.DecodeError()))
		return
	}

	validationError := validationx.Check(chart)
	if validationError != nil {
		loggerx.Error(errors.New("validation error"))
		errorx.Render(w, validationError)
		return
	}

	result := &model.Chart{}
	result.ID = uint(id)

	// check record exists or not
	err = config.DB.Where(&model.Chart{
		OrganisationID: uint(oID),
	}).First(&result).Error

	if err != nil {
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.RecordNotFound()))
		return
	}

	// Get table name
	stmt := &gorm.Statement{DB: config.DB}
	_ = stmt.Parse(&model.Chart{})
	tableName := stmt.Schema.Table

	var chartSlug string

	if result.Slug == chart.Slug {
		chartSlug = result.Slug
	} else if chart.Slug != "" && slug.Check(chart.Slug) {
		chartSlug = slug.Approve(chart.Slug, oID, tableName)
	} else {
		chartSlug = slug.Approve(slug.Make(chart.Title), oID, tableName)
	}

	tx := config.DB.Begin()

	newTags := make([]model.Tag, 0)
	if len(chart.TagIDs) > 0 {
		tx.Model(&model.Tag{}).Where(chart.TagIDs).Find(&newTags)
		if err = tx.Model(&result).Association("Tags").Replace(&newTags); err != nil {
			tx.Rollback()
			loggerx.Error(err)
			errorx.Render(w, errorx.Parser(errorx.DBError()))
			return
		}
	} else {
		_ = tx.Model(&result).Association("Tags").Clear()
	}

	newCategories := make([]model.Category, 0)
	if len(chart.CategoryIDs) > 0 {
		tx.Model(&model.Category{}).Where(chart.CategoryIDs).Find(&newCategories)
		if err = tx.Model(&result).Association("Categories").Replace(&newCategories); err != nil {
			tx.Rollback()
			loggerx.Error(err)
			errorx.Render(w, errorx.Parser(errorx.DBError()))
			return
		}
	} else {
		_ = tx.Model(&result).Association("Categories").Clear()
	}

	featuredMediumID := &chart.FeaturedMediumID
	result.FeaturedMediumID = &chart.FeaturedMediumID
	if chart.FeaturedMediumID == 0 {
		err = tx.Model(&result).Omit("Tags", "Categories").Updates(map[string]interface{}{"featured_medium_id": nil}).Error
		featuredMediumID = nil
		if err != nil {
			tx.Rollback()
			loggerx.Error(err)
			errorx.Render(w, errorx.Parser(errorx.DBError()))
			return
		}
	}

	themeID := &chart.ThemeID
	result.ThemeID = &chart.ThemeID
	if chart.ThemeID == 0 {
		err = tx.Model(&result).Omit("Tags", "Categories").Updates(map[string]interface{}{"theme_id": nil}).Error
		themeID = nil
		if err != nil {
			tx.Rollback()
			loggerx.Error(err)
			errorx.Render(w, errorx.Parser(errorx.DBError()))
			return
		}
	}

	err = tx.Model(&result).Omit("Tags", "Categories").Updates(model.Chart{
		Base:             config.Base{UpdatedByID: uint(uID)},
		Title:            chart.Title,
		Slug:             chartSlug,
		DataURL:          chart.DataURL,
		Description:      chart.Description,
		Status:           chart.Status,
		FeaturedMediumID: featuredMediumID,
		Config:           chart.Config,
		ThemeID:          themeID,
		PublishedDate:    chart.PublishedDate,
	}).Preload("Medium").Preload("Theme").Preload("Tags").Preload("Categories").First(&result).Error

	if err != nil {
		tx.Rollback()
		loggerx.Error(err)
		errorx.Render(w, errorx.Parser(errorx.DBError()))
		return
	}

	tx.Commit()
	renderx.JSON(w, http.StatusOK, result)
}
