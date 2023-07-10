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
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	*a = strings.Split(string(bytes), arraySeparator)
	return nil
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return strings.Join(a, arraySeparator), nil
}
