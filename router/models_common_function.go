package router

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

func GetList(db *gorm.DB, conditions *DbConditions) *Paginator {
	var total int64
	page := conditions.PaginatorParam.Page
	pageSize := conditions.PaginatorParam.PerPage

	base := dbAddConditions(db, conditions)

	base.Count(&total)

	base = base.Limit(pageSize).Offset((page - 1) * pageSize)
	base.Find(conditions.Model)

	pageNum := int(total) / pageSize
	if int(total)%pageSize != 0 {
		pageNum++
	}

	return &Paginator{
		Data:        conditions.Model,
		CurrentPage: page,
		PerPage:     strconv.Itoa(pageSize),
		Total:       int(total),
	}
}

func GetRecord(db *gorm.DB, conditions *DbConditions) (interface{}, error) {
	base := dbAddConditions(db, conditions)
	err := base.First(conditions.Model).Error
	return conditions.Model, err
}

func GetRecords(db *gorm.DB, conditions *DbConditions) interface{} {
	base := dbAddConditions(db, conditions)
	base.Find(conditions.Model)
	return conditions.Model
}

func CreateRecord(db *gorm.DB, record interface{}) (interface{}, error) {
	base := db
	err := base.Omit("created_at", "updated_at", "deleted_at").Create(record).Error
	return record, err
}

func DeleteRecord(db *gorm.DB, record interface{}, filters *Filters) (interface{}, error) {
	base := db
	base = dbAddFilters(base, filters)
	err := base.Delete(record).Error
	return record, err
}

func UpdateRecord(db *gorm.DB, records interface{}, omit []string, filters *Filters) error {
	base := db.Model(records)
	base = dbAddFilters(base, filters)

	if omit != nil {
		for _, o := range omit {
			base.Omit(o)
		}
	}

	err := base.Updates(records).Error
	return err
}

func UpdateRecordUseData(db *gorm.DB, model interface{}, omit []string, filters *Filters, data interface{}) error {
	base := db.Model(model)
	base = dbAddFilters(base, filters)

	if omit != nil {
		for _, o := range omit {
			base.Omit(o)
		}
	}

	err := base.Updates(data).Error
	return err
}

func dbAddFilters(db *gorm.DB, filters *Filters) *gorm.DB {
	base := db
	for _, f := range *filters {
		if f.Value() == nil {
			base = base.Where(f.Expression())
		} else {
			base = base.Where(f.Expression(), f.Value())
		}
	}
	return base
}

func dbAddConditions(db *gorm.DB, conditions *DbConditions) *gorm.DB {
	base := db.Model(conditions.Model)
	if conditions.Filters != nil {
		base = dbAddFilters(base, conditions.Filters)
	}
	if conditions.Order != nil {
		base = base.Order(conditions.Order.Value())
	}
	if conditions.Fields != nil {
		base = base.Select(conditions.Fields)
	}
	if conditions.GroupBy != "" {
		base = base.Group(conditions.GroupBy)
	}
	if conditions.Omit != nil {
		for _, o := range conditions.Omit {
			base = base.Omit(o)
		}
	}
	if conditions.Preloads != nil {
		for _, p := range conditions.Preloads {
			base = base.Preload(p)
		}
	}
	if conditions.PreloadsWithFunction != nil {
		for preload, preFunc := range conditions.PreloadsWithFunction {
			base = base.Preload(preload, preFunc)
		}
	}

	return base
}
