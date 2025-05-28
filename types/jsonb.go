package types

import (
	"database/sql/driver"
	"encoding/json"
)

type JsonB map[string]any

// Serializer manual para o GORM
func (m JsonB) Value() (driver.Value, error) { return json.Marshal(m) }

func (m *JsonB) Scan(vl any) error {
	if vl == nil {
		*m = JsonB{}
		return nil
	}
	return json.Unmarshal(vl.([]byte), m)
}
