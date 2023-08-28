package router

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type DbConditions struct {
	Omit                 []string
	Model                interface{}
	Filters              *Filters
	PaginatorParam       *PaginatorRequestParam
	Order                *Order
	GroupBy              string
	Fields               []string
	Preloads             []string
	PreloadsWithFunction map[string]func(db *gorm.DB) *gorm.DB
}

type Filter []interface{}

func NewFilter(expression string, value interface{}) Filter {
	return Filter{expression, value}
}
func (f *Filter) Expression() interface{} {
	fc := *f
	return fc[0]
}
func (f *Filter) Value() interface{} {
	fc := *f
	return fc[1]
}

type Filters []*Filter

func (fs *Filters) Add(expression string, value interface{}) Filters {
	f := NewFilter(expression, value)
	fc := append(*fs, &f)
	return fc
}

type Order struct {
	Order string
	Sort  string
}

func NewOrderBy(conditions ...string) Order {
	var order string
	var sort string
	for i, c := range conditions {
		if i == 0 {
			order = c
		}
		if i == 1 {
			sort = c
		}
		if i > 1 {
			break
		}
	}
	if order == "" {
		order = "created_at"
	}
	if sort == "" {
		sort = "desc"
	}
	return Order{
		Order: order,
		Sort:  sort,
	}
}

func (o *Order) Value() string {
	return fmt.Sprintf("%s %s", o.Order, o.Sort)
}
