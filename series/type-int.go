package series

import (
	"fmt"
	"math"
	"strconv"
)

type IntElement struct {
	e     int64 // https://golang.org/ref/spec#Numeric_types forcing to be int64
	valid bool
	nan   bool // bookkeeping
}

func (e *IntElement) Set(value interface{}) error {
	e.valid = true
	e.nan = false
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
			i, err := strconv.ParseInt(value.(string), 10, 64)
			if err != nil {
				e.nan = true
				return err
			}
			e.e = i
		}
	case int:
		e.e = int64(value.(int))
	case int8:
		e.e = int64(value.(int8))
	case int16:
		e.e = int64(value.(int16))
	case int32:
		e.e = int64(value.(int32))
	case int64:
		e.e = int64(value.(int64))
	case uint:
		e.e = int64(value.(uint))
	case uint8:
		e.e = int64(value.(uint8))
	case uint16:
		e.e = int64(value.(uint16))
	case uint32:
		e.e = int64(value.(uint32))
	case uint64:
		e.e = int64(value.(uint64))
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
		e.e = int64(f)
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
		e.e = int64(f)
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
			v, err := value.(Element).Int()
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
		return fmt.Errorf("Unsupported type '%T' conversion to an int64", value)
	}
	return nil
}

func (e IntElement) Copy() Element {
	return &IntElement{e: e.e, valid: e.valid, nan: e.nan}
}

func (e IntElement) IsNaN() bool {
	return !e.valid || e.nan
}

func (e IntElement) IsValid() bool {
	return e.valid
}

func (e IntElement) IsInf(sign int) bool {
	return false
}

func (e IntElement) Type() Type {
	return Int
}

func (e IntElement) Val() ElementValue {
	if e.valid {
		if e.nan {
			return NaNElement{}
		}
		return int(e.e)
	}
	return nil
}

func (e IntElement) String() (string, error) {
	if e.valid {
		if e.nan {
			return NaN, nil
		}
		return fmt.Sprint(e.e), nil
	}
	return Nil, nil
}

func (e IntElement) Int() (int64, error) {
	if e.valid && !e.nan {
		return e.e, nil
	}
	return 0, fmt.Errorf("can't convert nil/nan to int64")
}

func (e IntElement) Uint() (uint64, error) {
	if e.valid && !e.nan {
		return uint64(e.e), nil
	}
	return 0, fmt.Errorf("can't convert nil/nan to uint64")
}

func (e IntElement) Float() (float64, error) {
	if e.valid {
		if e.nan {
			return math.NaN(), nil
		}
		return float64(e.e), nil
	}
	return math.NaN(), fmt.Errorf("can't convert nil to float64")
}

func (e IntElement) Bool() (bool, error) {
	if !e.valid {
		return false, fmt.Errorf("can't convert nil to Bool")
	}
	if e.IsNaN() {
		return true, nil // not zero so true
	}
	if e.e == 0 {
		return false, nil
	}
	return true, nil
}

func (e IntElement) Eq(elem Element) bool {
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
	i, err := elem.Int()
	if err != nil {
		return false
	}
	return e.e == i
}

func (e IntElement) Neq(elem Element) bool {
	if e.valid != elem.IsValid() {
		return true
	}
	return !e.Eq(elem)
}

func (e IntElement) Less(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return true
	}
	i, err := elem.Int()
	if err != nil {
		return false
	}
	return e.e < i
}

func (e IntElement) LessEq(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return true
	}
	i, err := elem.Int()
	if err != nil || !e.IsValid() {
		return false
	}
	return e.e <= i
}

func (e IntElement) Greater(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return false
	}
	i, err := elem.Int()
	if err != nil {
		return false
	}
	return e.e > i
}

func (e IntElement) GreaterEq(elem Element) bool {
	if e.IsNaN() || elem.IsNaN() {
		return false
	}
	if elem.IsNaN() {
		return false
	}
	if elem.IsInf(1) {
		return false
	}
	i, err := elem.Int()
	if err != nil {
		return false
	}
	return e.e >= i
}
