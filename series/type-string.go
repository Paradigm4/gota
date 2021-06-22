package series

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type stringElement struct {
	e     string
	valid bool
}

// Strings with NaN will be treated as just strings with NaN
func (e *stringElement) Set(value interface{}) error {
	e.valid = true
	if value == nil {
		e.valid = false
		return nil
	}
	switch value.(type) {
	case string:
		if value.(string) == Nil {
			e.valid = false
		} else {
			e.e = string(value.(string))
		}
	case int:
		e.e = strconv.Itoa(value.(int))
	case int64:
		e.e = strconv.FormatInt(value.(int64), 10)
	case uint64:
		e.e = strconv.FormatUint(value.(uint64), 10)
	case float32:
		e.e = strconv.FormatFloat(float64(value.(float32)), 'f', 6, 64)
	case float64:
		e.e = strconv.FormatFloat(value.(float64), 'f', 6, 64)
	case bool:
		b := value.(bool)
		if b {
			e.e = "true"
		} else {
			e.e = "false"
		}
	case NaNElement:
		e.e = "NaN"
	case Element:
		if value.(Element).IsValid() {
			v, err := value.(Element).String()
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
		return fmt.Errorf("Unsupported type '%T' conversion to a string", value)
	}
	return nil
}

func (e stringElement) Copy() Element {
	return &stringElement{e.e, e.valid}
}

// Returns true if the string is parsed as NaN, missing, or fails to be parsed as a float
func (e stringElement) IsNaN() bool {
	if !e.valid {
		return true
	}
	f, err := strconv.ParseFloat(e.e, 64)
	if err == nil {
		return math.IsNaN(f)
	}
	return true
}

func (e stringElement) IsValid() bool {
	return e.valid
}

// Returns true if the string is parsed as Inf, -Inf or +Inf.
func (e stringElement) IsInf(sign int) bool {
	switch strings.ToLower(e.e) {
	case "inf", "-inf", "+inf":
		f, err := strconv.ParseFloat(e.e, 64)
		if err != nil {
			return false
		}
		return math.IsInf(f, sign)
	}
	return false
}

func (e stringElement) Type() Type {
	return String
}

func (e stringElement) Val() ElementValue {
	if !e.IsValid() {
		return nil
	}
	return string(e.e)
}

func (e stringElement) String() (string, error) {
	if !e.IsValid() {
		return "", fmt.Errorf("can't convert nil to string")
	}
	return string(e.e), nil
}

func (e stringElement) Int() (int64, error) {
	if !e.IsValid() {
		return 0, fmt.Errorf("can't convert nil to int64")
	}
	return strconv.ParseInt(e.e, 10, 64)
}

func (e stringElement) Uint() (uint64, error) {
	if !e.IsValid() {
		return 0, fmt.Errorf("can't convert nil to uint64")
	}
	return strconv.ParseUint(e.e, 10, 64)
}

func (e stringElement) Float() (float64, error) {
	if !e.IsValid() {
		return math.NaN(), fmt.Errorf("can't convert nil to float64")
	}
	f, err := strconv.ParseFloat(e.e, 64)
	if err != nil {
		return math.NaN(), nil
	}
	return f, nil
}

func (e stringElement) Bool() (bool, error) {
	if !e.IsValid() {
		return false, fmt.Errorf("can't convert nil to bool")
	}
	switch strings.ToLower(e.e) {
	case "true", "t", "1":
		return true, nil
	case "false", "f", "0":
		return false, nil
	}
	return false, fmt.Errorf("can't convert String '%v' to bool", e.e)
}

func (e stringElement) Eq(elem Element) bool {
	if e.valid != elem.IsValid() {
		// xor
		return false
	}
	if !e.valid && !elem.IsValid() {
		// nil == nil is true
		return true
	}
	s, err := elem.String()
	if err != nil {
		return false
	}
	return e.e == s
}

func (e stringElement) Neq(elem Element) bool {
	return !e.Eq(elem)
}

func (e stringElement) Less(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	s, err := elem.String()
	if err != nil {
		return false
	}
	return e.e < s
}

func (e stringElement) LessEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	s, err := elem.String()
	if err != nil {
		return false
	}
	return e.e <= s
}

func (e stringElement) Greater(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	s, err := elem.String()
	if err != nil {
		return false
	}
	return e.e > s
}

func (e stringElement) GreaterEq(elem Element) bool {
	if !e.valid || !elem.IsValid() {
		// really should be an error
		return false
	}
	s, err := elem.String()
	if err != nil {
		return false
	}
	return e.e >= s
}
