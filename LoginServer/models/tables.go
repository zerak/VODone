package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type Table struct {
	ID   int       "auto"
	Time time.Time "type(datetime)"
}

type Mtable struct {
	ID   int       "auto"
	Time time.Time "type(datetime)"
}

func init() {
	orm.RegisterModel(new(Table))
	orm.RegisterModel(new(Mtable))
}

