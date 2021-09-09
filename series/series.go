package series

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"math"

	"gonum.org/v1/gonum/stat"
)

// Series is a data structure designed for operating on arrays of elements that
// should comply with a certain type structure. They are flexible enough that can
// be transformed to other Series types and account for missing or non valid
// elements. Most of the power of Series resides on the ability to compare and
// subset Series of different types.
type Series struct {
	Name         string      // The name of the series
	elements     Elements    // The values of the elements
	defaultValue interface{} // value to use for nil, if nil then element will be set InValid
	t            Type        // The type of the series
	Err          error       // If there are errors they are stored here
}

// Elements is the interface that represents the array of elements contained on
// a Series.
type Elements interface {
	Elem(int) Element
	Len() int
}

// Element is the interface that defines the types of methods to be present for elements of a Series
type Element interface {
	// Setter method
	Set(interface{}) error

	// Information methods
	IsValid() bool  // if not valid, that value is missing (nil).
	IsNaN() bool    // test if value is a floating point NaN. Non-valid elements will return true but backing series is only guaranteed to return an NaN for Floats and Strings conversions
	IsInf(int) bool // same as math.IsInf
	Type() Type

	// Comparation methods
	Eq(Element) bool
	Neq(Element) bool
	Less(Element) bool
	LessEq(Element) bool
	Greater(Element) bool
	GreaterEq(Element) bool

	// Accessor/conversion methods
	Copy() Element     // FIXME: Returning interface is a recipe for pain
	Val() ElementValue // FIXME: Returning interface is a recipe for pain - can be used to test if NaN or just non-valid (nil)
	String() (string, error)
	Int() (int64, error)     // All signed ints are stored as int64s, but indices remain system dependent int type
	Uint() (uint64, error)   // All unsigned ints are stored as uint64s
	Float() (float64, error) // only returns NaN if valid, if not valid will error.
	Bool() (bool, error)
}

// Place holder for elements that are NaN. Complex rules for how this is promoted depending on the series type
type NaNElement struct{}

// intElements is the concrete implementation of Elements for Int elements.
type intElements []IntElement

func (e intElements) Len() int           { return len(e) }
func (e intElements) Elem(i int) Element { return &e[i] }

// uintElements is the concrete implementation of Elements for Int elements.
type uintElements []uintElement

func (e uintElements) Len() int           { return len(e) }
func (e uintElements) Elem(i int) Element { return &e[i] }

// stringElements is the concrete implementation of Elements for String elements.
type stringElements []stringElement

func (e stringElements) Len() int           { return len(e) }
func (e stringElements) Elem(i int) Element { return &e[i] }

// floatElements is the concrete implementation of Elements for Float elements.
type floatElements []floatElement

func (e floatElements) Len() int           { return len(e) }
func (e floatElements) Elem(i int) Element { return &e[i] }

// boolElements is the concrete implementation of Elements for Bool elements.
type boolElements []boolElement

func (e boolElements) Len() int           { return len(e) }
func (e boolElements) Elem(i int) Element { return &e[i] }

// ElementValue represents the value that can be used for marshaling or unmarshaling Elements.
type ElementValue interface{}

type MapFunction func(Element) Element

// Comparator is a convenience alias that can be used for a more type safe way of
// reason and use comparators.
type Comparator string

// Supported Comparators
const (
	Eq        Comparator = "==" // Equal
	Neq       Comparator = "!=" // Non equal
	Greater   Comparator = ">"  // Greater than
	GreaterEq Comparator = ">=" // Greater or equal than
	Less      Comparator = "<"  // Lesser than
	LessEq    Comparator = "<=" // Lesser or equal than
	In        Comparator = "in" // Inside
)

// Type is a convenience alias that can be used for a more type safe way of
// reason and use Series types.
type Type string

// Supported Series Types
// All Int-types are stored as int64
const (
	String  Type = "string"
	Int     Type = "int64"   // all signed ints are stored internally as int64
	Uint    Type = "uint64"  // all unsigned ints are stored internally as unit64s
	Float   Type = "float64" // same as Float64
	Float64 Type = "float64"
	Bool    Type = "bool"
	// not series types. these are string representations used for conversions
	NaN string = "NaN"
	Nil string = ""
)

// Indexes represent the elements that can be used for selecting a subset of
// elements within a Series. Currently supported are:
//
//     int            // Matches the given index number
//     []int          // Matches all given index numbers
//     []bool         // Matches all elements in a Series marked as true
//     Series [Int]   // Same as []int
//     Series [Bool]  // Same as []bool
type Indexes interface{}

// int max function
func imax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// New is the generic Series constructor
// For a non-empty series when passing a nil value, set size to 1
func New(values interface{}, t Type, name string, size ...int) Series {
	return NewDefault(values, nil, t, name, size...)
}

// NewDefault uses the given defaultValue to fill a series when a nil is encountered.
// For a non-empty series when passing a nil value, set size to 1
func NewDefault(values interface{}, defaultValue interface{}, t Type, name string, size ...int) Series {
	ret := Series{
		Name: name,
		t:    t,
		Err:  nil,
	}

	alloc_size := -1
	if size != nil && len(size) == 1 {
		alloc_size = size[0]
	}

	// Pre-allocate elements
	preAlloc := func(n int) {
		switch t {
		case String:
			ret.elements = make(stringElements, n)
		case Int:
			ret.elements = make(intElements, n)
		case Uint:
			ret.elements = make(uintElements, n)
		case Float:
			ret.elements = make(floatElements, n)
		case Bool:
			ret.elements = make(boolElements, n)
		default:
			panic(fmt.Sprintf("unknown type %v", t))
		}
	}

	if values == nil {
		l := imax(alloc_size, 0) // so we can create an empty DataFrame (no rows)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).Set(defaultValue)
		}
		return ret
	}

	switch values.(type) {
	case []string:
		v := values.([]string)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v[i])
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []float32:
		v := values.([]float32)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(float64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []float64:
		v := values.([]float64)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v[i])
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []int:
		v := values.([]int)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(int64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []int8:
		v := values.([]int8)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(int64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []int16:
		v := values.([]int16)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(int64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []int32:
		v := values.([]int32)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(int64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []int64:
		v := values.([]int64)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v[i])
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []uint:
		v := values.([]uint)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(uint64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []uint8:
		v := values.([]uint8)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(uint64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []uint16:
		v := values.([]uint16)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(uint64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []uint32:
		v := values.([]uint32)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(uint64(v[i]))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []uint64:
		v := values.([]uint64)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v[i])
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case []bool:
		v := values.([]bool)
		l := len(v)
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v[i])
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	case Series:
		v := values.(Series)
		l := v.Len()
		a := imax(l, alloc_size)
		preAlloc(a)
		for i := 0; i < a; i++ {
			if i < l {
				ret.elements.Elem(i).Set(v.elements.Elem(i))
			} else {
				ret.elements.Elem(i).Set(nil)
			}
		}
	default:
		switch reflect.TypeOf(values).Kind() {
		case reflect.Slice:
			v := reflect.ValueOf(values)
			l := v.Len()
			a := imax(l, alloc_size)
			preAlloc(a)
			for i := 0; i < a; i++ {
				if i < l {
					val := v.Index(i).Interface()
					if val == nil {
						ret.elements.Elem(i).Set(defaultValue)
					} else {
						ret.elements.Elem(i).Set(val)
					}
				} else {
					ret.elements.Elem(i).Set(nil)
				}
			}
		default:
			preAlloc(1)
			v := reflect.ValueOf(values)
			val := v.Interface()
			if val == nil {
				ret.elements.Elem(0).Set(defaultValue)
			} else {
				ret.elements.Elem(0).Set(val)
			}
		}
	}

	return ret
}

// Strings is a constructor for a String Series
func Strings(values interface{}) Series {
	return New(values, String, "")
}

// Ints is a constructor for an Int Series
func Ints(values interface{}) Series {
	return New(values, Int, "")
}

// Uints is a constructor for an Int Series
func Uints(values interface{}) Series {
	return New(values, Uint, "")
}

// Floats is a constructor for a Float Series
func Floats(values interface{}) Series {
	return New(values, Float, "")
}

// Bools is a constructor for a Bool Series
func Bools(values interface{}) Series {
	return New(values, Bool, "")
}

// Empty returns an empty Series of the same type
func (s Series) Empty() Series {
	return New([]int{}, s.t, s.Name)
}

// Append adds new elements to the end of the Series.
// When using Append, the Series is modified in place.
func (s *Series) Append(values interface{}) {
	// TODO: Need an AppendNull, currently Append(nil) inserts nothing
	if err := s.Err; err != nil {
		return
	}
	news := NewDefault(values, s.defaultValue, s.t, s.Name)
	switch s.t {
	case String:
		s.elements = append(s.elements.(stringElements), news.elements.(stringElements)...)
	case Int:
		s.elements = append(s.elements.(intElements), news.elements.(intElements)...)
	case Uint:
		s.elements = append(s.elements.(uintElements), news.elements.(uintElements)...)
	case Float:
		s.elements = append(s.elements.(floatElements), news.elements.(floatElements)...)
	case Bool:
		s.elements = append(s.elements.(boolElements), news.elements.(boolElements)...)
	default:
		panic(fmt.Sprintf("unknown type %v", s.t))
	}
}

// Concat concatenates two series together.
//
// It will return a new Series with the combined elements of both Series.
func (s Series) Concat(x Series) Series {
	if err := s.Err; err != nil {
		return s
	}
	if err := x.Err; err != nil {
		s.Err = fmt.Errorf("concat error: argument has errors: %v", err)
		return s
	}
	y := s.Copy()
	y.Append(x)
	return y
}

// Subset returns a subset of the series based on the given Indexes.
func (s Series) Subset(indexes Indexes) Series {
	if err := s.Err; err != nil {
		return s
	}
	idx, err := parseIndexes(s.Len(), indexes)
	if err != nil {
		s.Err = err
		return s
	}
	ret := Series{
		Name: s.Name,
		t:    s.t,
	}
	switch s.t {
	case String:
		elements := make(stringElements, len(idx))
		for k, i := range idx {
			elements[k] = s.elements.(stringElements)[i]
		}
		ret.elements = elements
	case Int:
		elements := make(intElements, len(idx))
		for k, i := range idx {
			elements[k] = s.elements.(intElements)[i]
		}
		ret.elements = elements
	case Uint:
		elements := make(uintElements, len(idx))
		for k, i := range idx {
			elements[k] = s.elements.(uintElements)[i]
		}
		ret.elements = elements
	case Float:
		elements := make(floatElements, len(idx))
		for k, i := range idx {
			elements[k] = s.elements.(floatElements)[i]
		}
		ret.elements = elements
	case Bool:
		elements := make(boolElements, len(idx))
		for k, i := range idx {
			elements[k] = s.elements.(boolElements)[i]
		}
		ret.elements = elements
	default:
		panic("unknown series type")
	}
	return ret
}

// TODO: should this be (s *Series) Set(index int, value interface{}) { ... }
func (s Series) Set(index int, value interface{}) Series {
	if err := s.Err; err != nil {
		return s
	}
	if index < 0 || index >= s.elements.Len() {
		s.Err = fmt.Errorf("index out of bounds: %d", index)
	}
	elem := s.elements.Elem(index)
	if value == nil {
		elem.Set(s.defaultValue)
	} else {
		elem.Set(value)
	}
	return s
}

// Update sets the values on the indexes of a Series and returns the reference
// for itself. The original Series is modified.
func (s Series) Update(indexes Indexes, newvalues Series) Series {
	if err := s.Err; err != nil {
		return s
	}
	if err := newvalues.Err; err != nil {
		s.Err = fmt.Errorf("set error: argument has errors: %v", err)
		return s
	}
	idx, err := parseIndexes(s.Len(), indexes)
	if err != nil {
		s.Err = err
		return s
	}
	if len(idx) != newvalues.Len() {
		s.Err = fmt.Errorf("set error: dimensions mismatch")
		return s
	}
	for k, i := range idx {
		if i < 0 || i >= s.Len() {
			s.Err = fmt.Errorf("set error: index out of range")
			return s
		}
		s.elements.Elem(i).Set(newvalues.elements.Elem(k))
	}
	return s
}

// HasInvalid checks whether the Series contains non-valid elements.
func (s Series) HasInvalid() bool {
	for i := 0; i < s.Len(); i++ {
		if !s.elements.Elem(i).IsValid() {
			return true
		}
	}
	return false
}

// Factorize  encode the Series as an enumerated type or categorical variable.
// This method is useful for obtaining a numeric representation of a Series
// when all that matters is identifying distinct values.
// Returns codes, uniques
func (s Series) Factorize(sort bool) (Series, Series) {
	// na_sentinel = interface{}

	err := s.Err
	if err != nil {
		codeSeries := New([]int{}, Int, "codes")
		codeSeries.Err = err
		unqiuesSeries := New([]interface{}{}, s.Type(), "uniques")
		unqiuesSeries.Err = err
		return codeSeries, unqiuesSeries
	}

	var sortedValues Series
	dropna := true
	codes := make([]int, s.Len())
	uniques := []interface{}{}
	uniquesMap := make(map[string]int) // because Go maps are a pain; can't use interface{} as a key.

	if sort {
		sortedValues = s.Subset(s.Order(false))
		if sortedValues.Err != nil {
			err = sortedValues.Err
			goto exit
		}
		// build uniquesMap in sorted order
		for i := 0; i < sortedValues.Len(); i++ {
			e := sortedValues.elements.Elem(i)
			if !((!e.IsValid() || (e.IsNaN() && e.Type() == Float)) && dropna) {
				se, err := e.String()
				if err != nil {
					goto exit
				}
				if u, ok := uniquesMap[se]; ok {
					codes[i] = u
				} else {
					u = len(uniques)
					uniques = append(uniques, e.Val())
					uniquesMap[se] = u
				}
			}
		}
	}
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		if (!e.IsValid() || (e.IsNaN() && e.Type() == Float)) && dropna {
			codes[i] = -1
		} else {
			se, err := e.String()
			if err != nil {
				goto exit
			}
			if u, ok := uniquesMap[se]; ok {
				codes[i] = u
			} else {
				u = len(uniques)
				uniques = append(uniques, e.Val())
				uniquesMap[se] = u
				codes[i] = u
			}
		}
	}

exit:
	if err != nil {
		codeSeries := New(codes, Int, "codes")
		codeSeries.Err = err
		unqiuesSeries := New(uniques, s.Type(), "uniques")
		unqiuesSeries.Err = err
		return codeSeries, unqiuesSeries
	}

	return New(codes, Int, "codes"), New(uniques, s.Type(), "uniques")
}

// HasNaN checks whether the Series contain NaN elements.
// These are elements that e.Float() would return as an NaN
func (s Series) HasNaN() bool {
	for i := 0; i < s.Len(); i++ {
		if s.elements.Elem(i).IsNaN() {
			return true
		}
	}
	return false
}

// IsNaN returns an array that identifies which of the elements are NaN.
func (s Series) IsNaN() []bool {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.elements.Elem(i).IsNaN()
	}
	return ret
}

// IsValid returns an array that identifies which of the elements are valid (not nil).
func (s Series) IsValid() []bool {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.elements.Elem(i).IsValid()
	}
	return ret
}

// Compare compares the values of a Series with other elements.
//
// To do so, the elements with are to be compared are first transformed to a
// Series of the same type as the caller.
func (s Series) Compare(comparator Comparator, comparando interface{}) Series {
	if err := s.Err; err != nil {
		return s
	}
	compareElements := func(a, b Element, c Comparator) (bool, error) {
		var ret bool
		switch c {
		case Eq:
			ret = a.Eq(b)
		case Neq:
			ret = a.Neq(b)
		case Greater:
			ret = a.Greater(b)
		case GreaterEq:
			ret = a.GreaterEq(b)
		case Less:
			ret = a.Less(b)
		case LessEq:
			ret = a.LessEq(b)
		default:
			return false, fmt.Errorf("unknown comparator: %v", c)
		}
		return ret, nil
	}

	comp := New(comparando, s.t, "")
	bools := make([]bool, s.Len())
	// In comparator comparation
	if comparator == In {
		for i := 0; i < s.Len(); i++ {
			e := s.elements.Elem(i)
			b := false
			for j := 0; j < comp.Len(); j++ {
				m := comp.elements.Elem(j)
				c, err := compareElements(e, m, Eq)
				if err != nil {
					s = s.Empty()
					s.Err = err
					return s
				}
				if c {
					b = true
					break
				}
			}
			bools[i] = b
		}
		return Bools(bools)
	}

	// Single element comparison
	if comp.Len() == 1 {
		for i := 0; i < s.Len(); i++ {
			e := s.elements.Elem(i)
			c, err := compareElements(e, comp.elements.Elem(0), comparator)
			if err != nil {
				s = s.Empty()
				s.Err = err
				return s
			}
			bools[i] = c
		}
		return Bools(bools)
	}

	// Multiple element comparison
	if s.Len() != comp.Len() {
		s := s.Empty()
		s.Err = fmt.Errorf("can't compare: length mismatch")
		return s
	}
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		c, err := compareElements(e, comp.elements.Elem(i), comparator)
		if err != nil {
			s = s.Empty()
			s.Err = err
			return s
		}
		bools[i] = c
	}
	return Bools(bools)
}

// Copy will return a copy of the Series.
func (s Series) Copy() Series {
	name := s.Name
	t := s.t
	err := s.Err
	var elements Elements
	switch s.t {
	case String:
		elements = make(stringElements, s.Len())
		copy(elements.(stringElements), s.elements.(stringElements))
	case Float:
		elements = make(floatElements, s.Len())
		copy(elements.(floatElements), s.elements.(floatElements))
	case Bool:
		elements = make(boolElements, s.Len())
		copy(elements.(boolElements), s.elements.(boolElements))
	case Int:
		elements = make(intElements, s.Len())
		copy(elements.(intElements), s.elements.(intElements))
	case Uint:
		elements = make(uintElements, s.Len())
		copy(elements.(uintElements), s.elements.(uintElements))
	default:
		panic(fmt.Sprintf("unsupported type %v", s.t))
	}
	ret := Series{
		Name:     name,
		t:        t,
		elements: elements,
		Err:      err,
	}
	return ret
}

// Records returns the elements of a Series as a []string
// If force is true and an element is not valid, an empty string will be inserted (promoted). Otherwise an error will be generated.
func (s Series) Records(force bool) ([]string, error) {
	ret := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.String()
		if err != nil && !force {
			return nil, err
		}
		if err != nil {
			ret[i] = ""
		} else {
			ret[i] = val
		}
	}
	return ret, nil
}

// Any returns whether any valid element is True
// Returns False unless there is at least one element within the series that is True or equivalent (e.g. non-zero or non-empty)
// This uses the element's Bool when not NA or Invalid
// If skipnan is false, then NaNs are treated as true; If skipnan is true and if the whole series is NaN, the result will be false
func (s Series) Any(skipnan bool) (bool, error) {
	rs := false
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		if !e.IsValid() {
			continue
		}
		if !skipnan && e.IsNaN() {
			rs = true
			break
		} else if skipnan && e.IsNaN() {
			continue
		} else {
			v, e := e.Bool()
			if e != nil {
				return rs, e
			}
			if v {
				rs = true
				break
			}
		}
	}
	return rs, nil
}

// Float returns the elements of a Series as a []float64.
// If foce is true and an element can not be converted to float64, an NaN will be inserted (promoted). Otherwise an error will be generated.
func (s Series) Float(force bool) ([]float64, error) {
	ret := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Float()
		if err != nil && !force {
			return nil, err
		}
		if err != nil {
			ret[i] = math.NaN()
		} else {
			ret[i] = val
		}
	}
	return ret, nil
}

// Int returns the elements of a Series as a []int64 or an error if the transformation is not possible.
func (s Series) Int() ([]int64, error) {
	ret := make([]int64, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Int()
		if err != nil {
			return nil, err
		}
		ret[i] = val
	}
	return ret, nil
}

// Uint returns the elements of a Series as a []uint64 or an error if the transformation is not possible.
func (s Series) Uint() ([]uint64, error) {
	ret := make([]uint64, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Uint()
		if err != nil {
			return nil, err
		}
		ret[i] = val
	}
	return ret, nil
}

// Bool returns the elements of a Series as a []bool or an error if the transformation is not possible.
func (s Series) Bool() ([]bool, error) {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Bool()
		if err != nil {
			return nil, err
		}
		ret[i] = val
	}
	return ret, nil
}

// Type returns the type of a given series
func (s Series) Type() Type {
	return s.t
}

// Len returns the length of a given Series
func (s Series) Len() int {
	return s.elements.Len()
}

// String
// This is only a representation and is not guaranteed to be able to recreate original series.
// Use Records for recreating the original series from a string representation
func (s Series) String() string {
	// TODO: limit output like DataFrame.String()
	ret := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		if e.IsValid() {
			s, err := e.String()
			if err != nil {
				ret[i] = ""
			} else {
				ret[i] = s
			}
		} else {
			ret[i] = ""
		}
	}
	return fmt.Sprint(ret)
}

// Str prints some extra information about a given series
func (s Series) Str() string {
	var ret []string
	// If name exists print name
	if s.Name != "" {
		ret = append(ret, "Name: "+s.Name)
	}
	ret = append(ret, "Type: "+fmt.Sprint(s.t))
	ret = append(ret, "Length: "+fmt.Sprint(s.Len()))
	if s.Len() != 0 {
		ret = append(ret, "Values: "+fmt.Sprint(s))
	}
	return strings.Join(ret, "\n")
}

// Val returns the value of a series for the given index.
// Will panic if the index is out of bounds.
func (s Series) Val(i int) interface{} {
	return s.elements.Elem(i).Val()
}

// Elem returns the element of a series for the given index.
// Will panic if the index is out of bounds.
func (s Series) Elem(i int) Element {
	return s.elements.Elem(i)
}

// parseIndexes will parse the given indexes for a given series of length `l`.
// No out of bounds checks is performed.
func parseIndexes(l int, indexes Indexes) ([]int, error) {
	var idx []int
	switch indexes.(type) {
	case []int:
		idx = indexes.([]int)
	case int:
		idx = []int{indexes.(int)}
	case []int64:
		ints := indexes.([]int64)
		for _, v := range ints {
			idx = append(idx, int(v))
		}
	case int64:
		idx = []int{int(indexes.(int64))}
	case []bool:
		bools := indexes.([]bool)
		if len(bools) != l {
			return nil, fmt.Errorf("indexing error: index dimensions mismatch")
		}
		for i, b := range bools {
			if b {
				idx = append(idx, int(i)) // idx is true
			}
		}
	case Series:
		s := indexes.(Series)
		if err := s.Err; err != nil {
			return nil, fmt.Errorf("indexing error: new values has errors: %v", err)
		}
		if s.HasNaN() {
			return nil, fmt.Errorf("indexing error: indexes contain NaN")
		}
		switch s.t {
		case Int:
			ints, err := s.Int()
			if err != nil {
				return nil, fmt.Errorf("indexing error: %v", err)
			}
			return parseIndexes(l, ints)
		case Bool:
			bools, err := s.Bool()
			if err != nil {
				return nil, fmt.Errorf("indexing error: %v", err)
			}
			return parseIndexes(l, bools)
		default:
			return nil, fmt.Errorf("indexing error: unknown indexing mode")
		}
	default:
		return nil, fmt.Errorf("indexing error: unknown indexing mode")
	}
	return idx, nil
}

// Order (stable-sort) starting with the given order.
func (s Series) OrderUsingIndex(reverse bool, origIdx []int) []int {
	var ie indexedElements
	var nasIdx []int
	if origIdx == nil || len(origIdx) == 0 {
		return s.Order(reverse)
	}
	slen := len(origIdx)
	if slen >= s.Len() {
		slen = s.Len()
	}
	for j := 0; j < slen; j++ {
		i := origIdx[j]
		e := s.elements.Elem(i)
		if !e.IsValid() {
			nasIdx = append(nasIdx, i)
		} else {
			switch s.t {
			case Float, Int, Uint:
				if e.IsNaN() {
					nasIdx = append(nasIdx, i)
				} else {
					ie = append(ie, indexedElement{i, e})
				}
			default:
				ie = append(ie, indexedElement{i, e})
			}
		}
	}
	var srt sort.Interface
	srt = ie
	if reverse {
		srt = sort.Reverse(srt)
	}
	sort.Sort(srt)
	var ret []int
	for _, e := range ie {
		ret = append(ret, e.index)
	}
	return append(ret, nasIdx...)
}

// Order returns the indexes for sorting a Series.
// NaN (and non-valid) elements are pushed to the end by order of appearance.
func (s Series) Order(reverse bool) []int {
	var ie indexedElements
	var nasIdx []int
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		if !e.IsValid() {
			nasIdx = append(nasIdx, i)
		} else {
			switch s.t {
			case Float, Int, Uint:
				if e.IsNaN() {
					nasIdx = append(nasIdx, i)
				} else {
					ie = append(ie, indexedElement{i, e})
				}
			default:
				ie = append(ie, indexedElement{i, e})
			}
		}
	}
	var srt sort.Interface
	srt = ie
	if reverse {
		srt = sort.Reverse(srt)
	}
	sort.Sort(srt)
	var ret []int
	for _, e := range ie {
		ret = append(ret, e.index)
	}
	return append(ret, nasIdx...)
}

type indexedElement struct {
	index   int
	element Element
}

type indexedElements []indexedElement

func (e indexedElements) Len() int           { return len(e) }
func (e indexedElements) Less(i, j int) bool { return e[i].element.Less(e[j].element) }
func (e indexedElements) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

// StdDev calculates the standard deviation of a series
func (s Series) StdDev() (float64, error) {
	vals, err := s.Float(false)
	if err != nil {
		return 0, err
	}
	stdDev := stat.StdDev(vals, nil)
	return stdDev, nil
}

// Mean calculates the average value of a series
func (s Series) Mean() (float64, error) {
	vals, err := s.Float(false)
	if err != nil {
		return 0, err
	}
	stdDev := stat.Mean(vals, nil)
	return stdDev, nil
}

// Median calculates the middle or median value, as opposed to
// mean, and there is less susceptible to being affected by outliers.
func (s Series) Median() (float64, error) {
	if s.elements.Len() == 0 ||
		s.Type() == String ||
		s.Type() == Bool {
		return math.NaN(), nil
	}
	ix := s.Order(false)
	newElem := make([]Element, len(ix))

	for newpos, oldpos := range ix {
		newElem[newpos] = s.elements.Elem(oldpos)
	}

	// When length is odd, we just take length(list)/2  value as the median.
	if len(newElem)%2 != 0 {
		return newElem[len(newElem)/2].Float()
	}
	// When length is even, we take middle two elements of list and the median is an average of the two of them.
	val1, err := newElem[(len(newElem)/2)-1].Float()
	if err != nil {
		return math.NaN(), err
	}
	val2, err := newElem[len(newElem)/2].Float()
	if err != nil {
		return math.NaN(), err
	}
	return (val1 + val2) * 0.5, nil
}

// Max return the biggest element in the series
func (s Series) Max() (float64, error) {
	if s.elements.Len() == 0 || s.Type() == String {
		return math.NaN(), nil
	}

	max := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Greater(max) {
			max = elem
		}
	}
	return max.Float()
}

// MaxStr return the biggest element in a series of type String
func (s Series) MaxStr() (string, error) {
	if s.elements.Len() == 0 || s.Type() != String {
		return "", nil
	}

	max := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Greater(max) {
			max = elem
		}
	}
	return max.String()
}

// Min return the lowest element in the series
func (s Series) Min() (float64, error) {
	if s.elements.Len() == 0 || s.Type() == String {
		return math.NaN(), nil
	}

	min := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Less(min) {
			min = elem
		}
	}
	return min.Float()
}

// MinStr return the lowest element in a series of type String
func (s Series) MinStr() (string, error) {
	if s.elements.Len() == 0 || s.Type() != String {
		return "", nil
	}

	min := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Less(min) {
			min = elem
		}
	}
	return min.String()
}

// Quantile returns the sample of x such that x is greater than or equal to the fraction p of samples.
// Note: gonum/stat panics when called with strings
func (s Series) Quantile(p float64) (float64, error) {
	if s.Type() == String || s.Len() == 0 {
		return math.NaN(), nil
	}

	ordered, err := s.Subset(s.Order(false)).Float(false)
	if err != nil {
		return math.NaN(), err
	}

	return stat.Quantile(p, stat.Empirical, ordered, nil), nil
}

// Map applies a function matching MapFunction signature, which itself
// allowing for a fairly flexible MAP implementation, intended for mapping
// the function over each element in Series and returning a new Series object.
// Function must be compatible with the underlying type of data in the Series.
// In other words it is expected that when working with a Float Series, that
// the function passed in via argument `f` will not expect another type, but
// instead expects to handle Element(s) of type Float.
func (s Series) Map(f MapFunction) Series {

	mappedValues := make([]Element, s.Len())
	for i := 0; i < s.Len(); i++ {
		value := f(s.elements.Elem(i))
		mappedValues[i] = value
	}
	return New(mappedValues, s.Type(), s.Name)
}
