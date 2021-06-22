package series

import (
	"fmt"
	"strings"
)

type boolElement struct {
	e     bool
	valid bool
}

func (e *boolElement) Set(value interface{}) error {
	e.valid = true
	e.e = false
	if value == nil {
		e.valid = false
		return nil
	}
	switch value.(type) {
	case string:
		switch strings.ToLower(value.(string)) {
		case "true", "t", "1":
			e.e = true
		case "false", "f", "0":
			e.e = false
		default:
			e.valid = false
			return fmt.Errorf("can't convert string '%s' to boolean", value.(string))
		}
	case int:
		if value.(int) == 0 {
			e.e = false
		} else {
			e.e = true
		}
	case int64:
		if value.(int64) == 0 {
			e.e = false
		} else {
			e.e = true
		}
	case uint:
		if value.(uint) == 0 {
			e.e = false
		} else {
			e.e = true
		}
	case uint64:
		if value.(uint64) == 0 {
			e.e = false
		} else {
			e.e = true
		}
	case float32:
		v := value.(float32)
		if v == 0 || v != v {
			e.e = false
		} else {
			e.e = true
		}
	case float64:
		v := value.(float64)
		if v == 0 || v != v {
			e.e = false
		} else {
			e.e = true
		}
	case bool:
		e.e = value.(bool)
	case Element:
		if value.(Element).IsValid() {
			b, err := value.(Element).Bool()
			if err != nil {
				e.valid = false
				return err
			}
			e.e = b
		} else {
			e.valid = false
			return nil
		}
	default:
		e.valid = false
		return fmt.Errorf("Unsupported type '%T' conversion to a boolean", value)
	}
	return nil
}

func (e boolElement) Copy() Element {
	return &boolElement{e.e, e.valid}
}

func (e boolElement) IsValid() bool {
	return e.valid
}

func (e boolElement) IsNaN() bool {
	if !e.valid {
		return true
	}
	return false
}

func (e boolElement) IsInf(sign int) bool {
	return false
}

func (e boolElement) Type() Type {
	return Bool
}

func (e boolElement) Val() ElementValue {
	if !e.valid {
		return nil
	}
	return bool(e.e)
}

func (e boolElement) String() (string, error) {
	if !e.valid {
		return "false", fmt.Errorf("can't convert a nil to string")
	}
	if e.e {
		return "true", nil
	}
	return "false", nil
}

func (e boolElement) Int() (int64, error) {
	if !e.valid {
		return 0, fmt.Errorf("can't convert a nil to an int64")
	}
	if e.e == true {
		return 1, nil
	}
	return 0, nil
}

func (e boolElement) Uint() (uint64, error) {
	if !e.valid {
		return 0, fmt.Errorf("can't convert a nil to an uint64")
	}
	if e.e == true {
		return 1, nil
	}
	return 0, nil
}

func (e boolElement) Float() (float64, error) {
	if !e.valid {
		return 0, fmt.Errorf("can't convert a nil to a float64")
	}
	if e.e {
		return 1.0, nil
	}
	return 0.0, nil
}

func (e boolElement) Bool() (bool, error) {
	if !e.valid {
		return false, fmt.Errorf("can't convert a nil to a boolean")
	}
	return bool(e.e), nil
}

func (e boolElement) Eq(elem Element) bool {
	if e.valid != elem.IsValid() {
		// xor
		return false
	}
	if !e.valid && !elem.IsValid() {
		// nil == nil is true
		return true
	}
	b, err := elem.Bool()
	if err != nil {
		return false
	}
	return e.e == b
}

func (e boolElement) Neq(elem Element) bool {
	if e.valid != elem.IsValid() {
		return true
	}
	return !e.Eq(elem)
}

func (e boolElement) Less(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if elem.IsNaN() {
		return false
	}
	b, err := elem.Bool()
	if err != nil {
		return false
	}
	return !e.e && b
}

func (e boolElement) LessEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if elem.IsNaN() {
		return false
	}
	b, err := elem.Bool()
	if err != nil {
		return false
	}
	return !e.e || b
}

func (e boolElement) Greater(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if elem.IsNaN() {
		return false
	}
	b, err := elem.Bool()
	if err != nil {
		return false
	}
	return e.e && !b
}

func (e boolElement) GreaterEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if elem.IsNaN() {
		return false
	}
	b, err := elem.Bool()
	if err != nil {
		return false
	}
	return e.e || !b
}
