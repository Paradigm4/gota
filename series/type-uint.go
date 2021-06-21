package series

import (
	"fmt"
	"math"
	"strconv"
)

type uintElement struct {
	e     uint64 // https://golang.org/ref/spec#Numeric_types forcing to be uint64
	valid bool
	nan   bool
}

func (e *uintElement) Set(value interface{}) error {
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
			e.nan = true
			return nil
		default:
			i, err := strconv.ParseUint(value.(string), 10, 64)
			if err != nil {
				e.nan = true
				return err
			}
			e.e = i
		}
	case int:
		e.e = uint64(value.(int))
	case int8:
		e.e = uint64(value.(int8))
	case int16:
		e.e = uint64(value.(int16))
	case int32:
		e.e = uint64(value.(int32))
	case int64:
		e.e = uint64(value.(int64))
	case uint:
		e.e = uint64(value.(uint))
	case uint8:
		e.e = uint64(value.(uint8))
	case uint16:
		e.e = uint64(value.(uint16))
	case uint32:
		e.e = uint64(value.(uint32))
	case uint64:
		e.e = uint64(value.(uint64))
	case float32:
		f := float64(value.(float32))
		if math.IsNaN(f) {
			e.nan = true
			return nil
		}
		if math.IsInf(f, 0) || math.IsInf(f, 1) {
			e.nan = true
			return fmt.Errorf("demoting float Inf to NaN")
		}
		e.e = uint64(f)
	case float64:
		f := value.(float64)
		if math.IsNaN(f) {
			e.nan = true
			return nil
		}
		if math.IsInf(f, 0) || math.IsInf(f, 1) {
			e.nan = true
			return fmt.Errorf("demoting float Inf to NaN")
		}
		e.e = uint64(f)
	case bool:
		b := value.(bool)
		if b {
			e.e = 1
		} else {
			e.e = 0
		}
	case NaNElement:
		e.nan = true
	case Element:
		if value.(Element).IsValid() {
			if value.(Element).IsNaN() {
				e.nan = true
				return nil
			}
			v, err := value.(Element).Uint()
			if err != nil {
				e.valid = false
				return err
			}
			e.e = v
		} else {
			e.valid = false
			return nil
		}
	default:
		e.valid = false
		return fmt.Errorf("Unsupported type '%T' conversion to an uint64", value)
	}
	return nil
}

func (e uintElement) Copy() Element {
	return &uintElement{e.e, e.valid, e.nan}
}

func (e uintElement) IsNaN() bool {
	return !e.valid || e.nan
}

func (e uintElement) IsValid() bool {
	return e.valid
}

func (e uintElement) IsInf(sign int) bool {
	return false
}

func (e uintElement) Type() Type {
	return Uint
}

func (e uintElement) Val() ElementValue {
	if e.valid {
		if e.nan {
			return NaNElement{}
		}
		return int(e.e)
	}
	return nil
}

func (e uintElement) String() (string, error) {
	if e.valid {
		if e.nan {
			return NaN, nil
		}
		return fmt.Sprint(e.e), nil
	}
	return Nil, nil
}

func (e uintElement) Int() (int64, error) {
	if e.valid && !e.nan {
		return int64(e.e), nil
	}
	return 0, fmt.Errorf("can't convert nil/nan to int64")
}

func (e uintElement) Uint() (uint64, error) {
	if e.valid && !e.nan {
		return uint64(e.e), nil
	}
	return 0, fmt.Errorf("can't convert nil/nan to uint64")
}

func (e uintElement) Float() (float64, error) {
	if e.valid {
		if e.nan {
			return math.NaN(), nil
		}
		return float64(e.e), nil
	}
	return math.NaN(), fmt.Errorf("can't convert nil to float64")
}

func (e uintElement) Bool() (bool, error) {
	if !e.valid || e.nan {
		return false, fmt.Errorf("can't convert nil/nan to bool")
	}
	switch e.e {
	case 1:
		return true, nil
	case 0:
		return false, nil
	}
	return false, fmt.Errorf("can't convert Int \"%v\" to bool", e.e)
}

func (e uintElement) Eq(elem Element) bool {
	if e.valid != elem.IsValid() {
		// xor
		return false
	}
	if !e.valid && !elem.IsValid() {
		// nil == nil is true
		return true
	}
	if elem.IsInf(0) {
		return false
	}
	i, err := elem.Uint()
	if err != nil {
		return false
	}
	return e.e == i
}

func (e uintElement) Neq(elem Element) bool {
	if e.valid != elem.IsValid() {
		return true
	}
	return !e.Eq(elem)
}

func (e uintElement) Less(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return true
	}
	i, err := elem.Uint()
	if err != nil {
		return false
	}
	return e.e < i
}

func (e uintElement) LessEq(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return true
	}
	i, err := elem.Uint()
	if err != nil || !e.IsValid() {
		return false
	}
	return e.e <= i
}

func (e uintElement) Greater(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return false
	}
	i, err := elem.Uint()
	if err != nil {
		return false
	}
	return e.e > i
}

func (e uintElement) GreaterEq(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return false
	}
	i, err := elem.Uint()
	if err != nil {
		return false
	}
	return e.e >= i
}
