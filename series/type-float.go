package series

import (
	"fmt"
	"math"
	"strconv"
)

type floatElement struct {
	e     float64
	valid bool
}

func (e *floatElement) Set(value interface{}) error {
	e.valid = true
	if value == nil {
		e.valid = false
		return nil
	}
	switch value.(type) {
	case string:
		switch value.(string) {
		case Nil:
			e.valid = false
			return nil
		case NaN:
			e.e = math.NaN()
			return nil
		default:
			f, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				e.e = math.NaN()
				return nil
			}
			e.e = f
		}
	case int:
		e.e = float64(value.(int))
	case int64:
		e.e = float64(value.(int64))
	case uint:
		e.e = float64(value.(uint))
	case uint64:
		e.e = float64(value.(uint64))
	case float32:
		e.e = float64(value.(float32))
	case float64:
		e.e = float64(value.(float64))
	case bool:
		b := value.(bool)
		if b {
			e.e = 1
		} else {
			e.e = 0
		}
	case Element:
		if value.(Element).IsValid() {
			f, err := value.(Element).Float()
			if err != nil {
				e.valid = false
				return err
			}
			e.e = f
		} else {
			e.valid = false
			return nil
		}
	default:
		e.valid = false
		return fmt.Errorf("Unsupported type '%T' conversion to a float", value)
	}
	return nil
}

func (e floatElement) Copy() Element {
	return &floatElement{e.e, e.valid}
}

func (e floatElement) IsValid() bool {
	return e.valid
}

func (e floatElement) IsNaN() bool {
	if !e.valid || math.IsNaN(e.e) {
		return true
	}
	return false
}

func (e floatElement) IsInf(sign int) bool {
	if e.valid {
		return math.IsInf(e.e, sign)
	}
	return false
}

func (e floatElement) Type() Type {
	return Float
}

func (e floatElement) Val() ElementValue {
	if !e.IsValid() {
		return nil
	}
	return float64(e.e)
}

func (e floatElement) String() (string, error) {
	if !e.IsValid() {
		return "", nil
	}
	return fmt.Sprintf("%f", e.e), nil
}

func (e floatElement) Int() (int64, error) {
	if !e.IsValid() {
		return 0, fmt.Errorf("can't convert nil to an int64")
	}
	f := e.e
	if math.IsInf(f, 1) || math.IsInf(f, -1) {
		return 0, fmt.Errorf("can't convert Inf to int64")
	}
	if math.IsNaN(f) {
		return 0, fmt.Errorf("can't convert NaN to int64")
	}
	return int64(f), nil
}

func (e floatElement) Uint() (uint64, error) {
	if !e.IsValid() {
		return 0, fmt.Errorf("can't convert nil to an uint64")
	}
	f := e.e
	if math.IsInf(f, 1) || math.IsInf(f, -1) {
		return 0, fmt.Errorf("can't convert Inf to uint64")
	}
	if math.IsNaN(f) {
		return 0, fmt.Errorf("can't convert NaN to uint64")
	}
	return uint64(f), nil
}

func (e floatElement) Float() (float64, error) {
	if !e.IsValid() {
		return 0, fmt.Errorf("can't convert nil to a float64")
	}
	return float64(e.e), nil
}

func (e floatElement) Bool() (bool, error) {
	if !e.IsValid() {
		return false, fmt.Errorf("can't convert nil to bool")
	}
	switch e.e {
	case 1:
		return true, nil
	case 0:
		return false, nil
	}
	return false, fmt.Errorf("can't convert Float '%v' to bool", e.e)
}

func (e floatElement) Eq(elem Element) bool {
	// NaN == NaN results to false in every language as per the IEEE 754
	if !e.valid && !elem.IsValid() {
		// both are not valid (nil)
		return true
	} else if e.valid != elem.IsValid() {
		// one is not valid
		return false
	}
	f, err := elem.Float()
	if err != nil {
		return false
	}
	return e.e == f
}

func (e floatElement) Neq(elem Element) bool {
	if e.valid != elem.IsValid() {
		return true
	}
	return !e.Eq(elem)
}

func (e floatElement) Less(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	f, err := elem.Float()
	if err != nil {
		return false
	}
	return e.e < f
}

func (e floatElement) LessEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	f, err := elem.Float()
	if err != nil {
		return false
	}
	return e.e <= f
}

func (e floatElement) Greater(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	f, err := elem.Float()
	if err != nil {
		return false
	}
	return e.e > f
}

func (e floatElement) GreaterEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	f, err := elem.Float()
	if err != nil {
		return false
	}
	return e.e >= f
}
