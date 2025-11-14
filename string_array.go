package fuse

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// StringArray is a wrapper around []string that implements the sql.Scanner and driver.Valuer interfaces
type StringArray []string

const (
	// arraySeparator is the string used to separate array elements in the database
	arraySeparator = ";"
)

func (a *StringArray) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*a = strings.Split(string(v), arraySeparator)
		return nil
	case string:
		*a = strings.Split(v, arraySeparator)
		return nil
	}
	return errors.New("src value cannot cast to []byte or string")
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return strings.Join(a, arraySeparator), nil
}
