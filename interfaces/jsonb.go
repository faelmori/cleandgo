package interfaces

import "database/sql/driver"

type IJsonB interface {
	Value() (driver.Value, error)
	Scan(vl any) error
}
