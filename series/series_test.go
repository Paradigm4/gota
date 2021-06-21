package series

// need to revisit these test cases for Record(force) and Float(force)
// add/update tests for HasNaN, HasInvalid, IsNaN, IsInvalid

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

// Check that there are no shared memory addreses between the elements of two Series
//func checkAddr(addra, addrb []string) error {
//for i := 0; i < len(addra); i++ {
//for j := 0; j < len(addrb); j++ {
//if addra[i] == "<nil>" || addrb[j] == "<nil>" {
//continue
//}
//if addra[i] == addrb[j] {
//return fmt.Errorf("found same address on\nA:%v\nB:%v", i, j)
//}
//}
//}
//return nil
//}

// Check that all the types on a Series are the same type and that it matches with
// Series.t
func checkTypes(s Series) error {
	var types []Type
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		types = append(types, e.Type())
	}
	for _, t := range types {
		if t != s.t {
			return fmt.Errorf("bad types for %v Series:\n%v", s.t, types)
		}
	}
	return nil
}

// compareFloats compares floating point values up to the number of digits specified.
// Returns true if both values are equal with the given precision
func compareFloats(lvalue, rvalue float64, digits int) bool {
	if math.IsNaN(lvalue) || math.IsNaN(rvalue) {
		return math.IsNaN(lvalue) && math.IsNaN(rvalue)
	}
	d := math.Pow(10.0, float64(digits))
	lv := int(lvalue * d)
	rv := int(rvalue * d)
	return lv == rv
}

// compare Eq, Neq, and In
func TestSeries_Compare_Eq(t *testing.T) {
	table := []struct {
		series     Series
		comparator Comparator
		comparando interface{}
		expected   Series
	}{
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Eq,
			"B",
			Bools([]bool{false, true, false, true, false, false}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Eq,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{false, true, true, false, false, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Eq,
			"2",
			Bools([]bool{false, true, false, false, false}),
		},
		{
			Ints([]string{"", "2", "1", "5", "NaN"}),
			Eq,
			"2",
			Bools([]bool{false, true, false, false, false}),
		},
		{
			Ints([]string{"", "2", "1", "5", "NaN"}),
			Eq,
			"",
			Bools([]bool{true, false, false, false, false}),
		},
		{
			Ints([]string{"", "2", "1", "5", "NaN"}),
			Eq,
			"NaN",
			Bools([]bool{false, false, false, false, false}), // NaN != NaN
		},
		{
			Strings(Ints([]string{"", "2", "1", "5", "NaN"})),
			Eq,
			"NaN",
			Bools([]bool{false, false, false, false, true}), // "NaN" == "NaN"
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Eq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, false}),
		},
		{
			New([]int{0, 2, 1, 5, 9}, Int, ""),
			Eq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, false}),
		},
		{
			New([]int{0, 2, 1, 5, 9}, Int, "", 6),
			Eq,
			[]int{0, 2, 0, 5, 10, -1},
			Bools([]bool{true, true, false, true, false, false}),
		},
		{
			New([]uint{0, 2, 1, 5, 9}, Uint, ""),
			Eq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, false}),
		},
		{
			New([]int{0, 2, 1, -5, 9}, Uint, ""), // -5 converted to uint64 is 18446744073709551611
			Eq,
			[]int{0, 2, 0, -5, 10},
			Bools([]bool{true, true, false, true, false}),
		},
		{
			New([]int{0, 2, 1, -5, 9}, Uint, ""),
			Eq,
			[]uint{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, false, false}),
		},
		{
			New([]uint8{0, 2, 1, 5, 9}, Uint, "", 6),
			Eq,
			[]int{0, 2, 0, 5, 10, -1},
			Bools([]bool{true, true, false, true, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Eq,
			"2",
			Bools([]bool{false, true, false, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Eq,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, false}),
		},
		{
			Bools([]bool{true, true, false}),
			Eq,
			"true",
			Bools([]bool{true, true, false}),
		},
		{
			Bools([]bool{true, true, false}),
			Eq,
			[]bool{true, false, false},
			Bools([]bool{true, false, true}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Neq,
			"B",
			Bools([]bool{true, false, true, false, true, true}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Neq,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{true, false, false, true, true, true}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Neq,
			"2",
			Bools([]bool{true, false, true, true, true}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Neq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, true}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Neq,
			"2",
			Bools([]bool{true, false, true, true, true}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Neq,
			[]uint{0, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Neq,
			"2",
			Bools([]bool{true, false, true, true, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Neq,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, true}),
		},
		{
			Bools([]bool{true, true, false}),
			Neq,
			"true",
			Bools([]bool{false, false, true}),
		},
		{
			Bools([]bool{true, true, false}),
			Neq,
			[]bool{true, false, false},
			Bools([]bool{false, true, false}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			In,
			"B",
			Bools([]bool{false, true, false, true, false, false}),
		},
		{
			Strings([]string{"Hello", "world", "this", "is", "a", "test"}),
			In,
			[]string{"cat", "world", "hello", "a"},
			Bools([]bool{false, true, false, false, true, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			In,
			"2",
			Bools([]bool{false, true, false, false, false}),
		},
		{
			Ints([]string{"", "2", "1", "5", "NaN"}),
			In,
			"",
			Bools([]bool{true, false, false, false, false}),
		},
		{
			Ints([]string{"", "2", "1", "5", "NaN"}),
			In,
			"NaN",
			Bools([]bool{false, false, false, false, false}), // NaN != NaN
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			In,
			[]int{2, 99, 1234, 9},
			Bools([]bool{false, true, false, false, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			In,
			"2",
			Bools([]bool{false, true, false, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			In,
			[]float64{2, 99, 1234, 9},
			Bools([]bool{false, true, false, false, true}),
		},
		{
			Bools([]bool{true, true, false}),
			In,
			"true",
			Bools([]bool{true, true, false}),
		},
		{
			Bools([]bool{true, true, false}),
			In,
			[]bool{false, false, false},
			Bools([]bool{false, false, true}),
		},
	}
	for testnum, test := range table {
		a := test.series
		b := a.Compare(test.comparator, test.comparando)
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected, _ := test.expected.Records(false)
		received, _ := b.Records(false)
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf(
				"Test:%v\nError:%v",
				testnum, err,
			)
		}
		//if err := checkAddr(a.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nA:%v\nB:%v", testnum, err, a.Addr(), b.Addr())
		//}
	}
}

func TestSeries_Compare_Greater(t *testing.T) {
	table := []struct {
		series     Series
		comparator Comparator
		comparando interface{}
		expected   Series
	}{
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Greater,
			"B",
			Bools([]bool{false, false, true, false, true, true}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Greater,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{false, false, false, false, true, true}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Greater,
			"2",
			Bools([]bool{false, false, false, true, true}),
		},
		{
			Ints([]string{"0", "2", "<nil>", "5", "A"}),
			Greater,
			"2",
			Bools([]bool{false, false, false, true, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Greater,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, false}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Greater,
			"2",
			Bools([]bool{false, false, false, true, true}),
		},
		{
			Uints([]string{"0", "2", "<nil>", "5", "A"}),
			Greater,
			"2",
			Bools([]bool{false, false, false, true, false}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Greater,
			[]uint{0, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Greater,
			"2",
			Bools([]bool{false, false, false, true, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Greater,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{false, false, true, false, false}),
		},
		{
			Bools([]bool{true, true, false}),
			Greater,
			"true",
			Bools([]bool{false, false, false}),
		},
		{
			Bools([]bool{true, true, false}),
			Greater,
			[]bool{true, false, false},
			Bools([]bool{false, true, false}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			GreaterEq,
			"B",
			Bools([]bool{false, true, true, true, true, true}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			GreaterEq,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{false, true, true, false, true, true}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			GreaterEq,
			"2",
			Bools([]bool{false, true, false, true, true}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			GreaterEq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, true, true, false}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			GreaterEq,
			"2",
			Bools([]bool{false, true, false, true, true}),
		},
		{
			Uints([]int{0, 2, 1, 5, 9}),
			GreaterEq,
			[]uint{0, 2, 0, 5, 10},
			Bools([]bool{true, true, true, true, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			GreaterEq,
			"2",
			Bools([]bool{false, true, false, true, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			GreaterEq,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{true, true, true, true, false}),
		},
		{
			Bools([]bool{true, true, false}),
			GreaterEq,
			"true",
			Bools([]bool{true, true, false}),
		},
		{
			Bools([]bool{true, true, false}),
			GreaterEq,
			[]bool{true, false, false},
			Bools([]bool{true, true, true}),
		},
	}
	for testnum, test := range table {
		a := test.series
		b := a.Compare(test.comparator, test.comparando)
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected, _ := test.expected.Records(false)
		received, _ := b.Records(false)
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf(
				"Test:%v\nError:%v",
				testnum, err,
			)
		}
	}
}

func TestSeries_Compare_Less(t *testing.T) {
	table := []struct {
		series     Series
		comparator Comparator
		comparando interface{}
		expected   Series
	}{
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Less,
			"B",
			Bools([]bool{true, false, false, false, false, false}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			Less,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{true, false, false, true, false, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Less,
			"2",
			Bools([]bool{true, false, true, false, false}),
		},
		{
			Ints([]string{"0", "2", "<nil>", "5", "A"}),
			Less,
			"2",
			Bools([]bool{true, false, false, false, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			Less,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{false, false, false, false, true}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Less,
			"2",
			Bools([]bool{true, false, true, false, false}),
		},
		{
			Uints([]string{"0", "2", "<nil>", "5", "A"}),
			Less,
			"2",
			Bools([]bool{true, false, false, false, false}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			Less,
			[]uint{0, 2, 0, 5, 10},
			Bools([]bool{false, false, false, false, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Less,
			"2",
			Bools([]bool{true, false, true, false, false}),
		},
		{
			Floats([]string{"0.1", "<nil>", "1", "5", "A"}),
			Less,
			"2",
			Bools([]bool{true, false, true, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			Less,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{false, false, false, false, true}),
		},
		{
			Bools([]bool{true, true, false}),
			Less,
			"true",
			Bools([]bool{false, false, true}),
		},
		{
			Bools([]bool{true, true, false}),
			Less,
			[]bool{true, false, false},
			Bools([]bool{false, false, false}),
		}, // need to add test for <nil>
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			LessEq,
			"B",
			Bools([]bool{true, true, false, true, false, false}),
		},
		{
			Strings([]string{"A", "B", "C", "B", "D", "BADA"}),
			LessEq,
			[]string{"B", "B", "C", "D", "A", "A"},
			Bools([]bool{true, true, true, true, false, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			LessEq,
			"2",
			Bools([]bool{true, true, true, false, false}),
		},
		{
			Ints([]string{"0", "2", "<nil>", "5", "A"}),
			LessEq,
			"2",
			Bools([]bool{true, true, false, false, false}),
		},
		{
			Ints([]int{0, 2, 1, 5, 9}),
			LessEq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, true}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			LessEq,
			"2",
			Bools([]bool{true, true, true, false, false}),
		},
		{
			Uints([]uint{0, 2, 1, 5, 9}),
			LessEq,
			[]int{0, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, true}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			LessEq,
			"2",
			Bools([]bool{true, true, true, false, false}),
		},
		{
			Floats([]string{"0.1", "2", "<nil>", "5", "A"}),
			LessEq,
			"2",
			Bools([]bool{true, true, false, false, false}),
		},
		{
			Floats([]float64{0.1, 2, 1, 5, 9}),
			LessEq,
			[]float64{0.1, 2, 0, 5, 10},
			Bools([]bool{true, true, false, true, true}),
		},
		{
			Bools([]bool{true, true, false}),
			LessEq,
			"true",
			Bools([]bool{true, true, true}),
		},
		{
			Bools([]bool{true, true, false}),
			LessEq,
			[]bool{true, false, false},
			Bools([]bool{true, false, true}),
		},
	}
	for testnum, test := range table {
		a := test.series
		b := a.Compare(test.comparator, test.comparando)
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected, _ := test.expected.Records(false)
		received, _ := b.Records(false)
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf(
				"Test:%v\nError:%v",
				testnum, err,
			)
		}
	}
}

func TestSeries_Subset(t *testing.T) {
	table := []struct {
		series   Series
		indexes  Indexes
		expected string
	}{
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			[]int{2, 1, 4, 4, 0, 3},
			"[C B D D A K]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			int(1),
			"[B]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			[]bool{true, false, false, true, true},
			"[A K D]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Ints([]int{3, 2, 1, 0}),
			"[K C B A]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Ints([]int{1}),
			"[B]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Ints(2),
			"[C]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Bools([]bool{true, false, false, true, true}),
			"[A K D]",
		},
	}
	for testnum, test := range table {
		a := test.series
		b := a.Subset(test.indexes)
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(b)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf(
				"Test:%v\nError:%v",
				testnum, err,
			)
		}
		//if err := checkAddr(a.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nA:%v\nB:%v", testnum, err, a.Addr(), b.Addr())
		//}
	}
}

// TODO TestSeries_Set(t *testing.T)
func TestSeries_Update(t *testing.T) {
	table := []struct {
		series   Series
		indexes  Indexes
		values   Series
		expected string
	}{
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			[]int{1, 2, 4},
			Ints([]string{"1", "2", "3"}),
			"[A 1 2 K 3]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			[]bool{false, true, true, false, true},
			Ints([]string{"1", "2", "3"}),
			"[A 1 2 K 3]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Ints([]int{1, 2, 4}),
			Ints([]string{"1", "2", "3"}),
			"[A 1 2 K 3]",
		},
		{
			Strings([]string{"A", "B", "C", "K", "D"}),
			Bools([]bool{false, true, true, false, true}),
			Ints([]string{"1", "2", "3"}),
			"[A 1 2 K 3]",
		},
	}
	for testnum, test := range table {
		b := test.series.Update(test.indexes, test.values)
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(b)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf(
				"Test:%v\nError:%v",
				testnum, err,
			)
		}
		//if err := checkAddr(test.values.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nNV:%v\nB:%v", testnum, err, test.values.Addr(), b.Addr())
		//}
	}
}

func TestStrings(t *testing.T) {
	table := []struct {
		series   Series
		expected string
	}{
		{
			Strings([]string{"A", "B", "C", "D"}),
			"[A B C D]",
		},
		{
			Strings([]string{"A", "B", "C", ""}),
			"[A B C ]",
		},
		{
			Strings([]string{"COL.3", "COL.1"}),
			"[COL.3 COL.1]",
		},
		{
			Strings([]string{"A"}),
			"[A]",
		},
		{
			Strings("A"),
			"[A]",
		},
		{
			Strings([]int{1, 2, 3}),
			"[1 2 3]",
		},
		{
			Strings([]int{2}),
			"[2]",
		},
		{
			Strings(-1),
			"[-1]",
		},
		{
			Strings([]float64{1, 2, 3}),
			"[1.000000 2.000000 3.000000]",
		},
		{
			Strings([]float64{2}),
			"[2.000000]",
		},
		{
			Strings(-1.0),
			"[-1.000000]",
		},
		{
			Strings(math.NaN()),
			"[NaN]",
		},
		{
			Strings(math.Inf(1)),
			"[+Inf]",
		},
		{
			Strings(math.Inf(-1)),
			"[-Inf]",
		},
		{
			Strings([]bool{true, true, false}),
			"[true true false]",
		},
		{
			Strings([]bool{false}),
			"[false]",
		},
		{
			Strings(true),
			"[true]",
		},
		{
			Strings([]int{}),
			"[]",
		},
		{
			Strings(nil),
			"[]",
		},
		{
			Strings(Strings([]string{"A", "B", "C"})),
			"[A B C]",
		},
	}
	for testnum, test := range table {
		if err := test.series.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(test.series.String())
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(test.series); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
	}
}

func TestInts(t *testing.T) {
	table := []struct {
		series   Series
		expected string
	}{
		{
			Ints([]string{"A", "B", "1", "2"}),
			"[NaN NaN 1 2]",
		},
		{
			Ints([]string{"A", "", "1", "2"}),
			"[NaN  1 2]",
		},
		{
			New(Ints([]string{"A", "", "1", "2"}), Int, ""),
			"[NaN  1 2]",
		},
		{
			Ints([]string{"1"}),
			"[1]",
		},
		{
			Ints("2"),
			"[2]",
		},
		{
			Ints([]int{1, 2, 3}),
			"[1 2 3]",
		},
		{
			Ints([]int{2}),
			"[2]",
		},
		{
			Ints(-1),
			"[-1]",
		},
		{
			Ints([]float64{1, 2, 3}),
			"[1 2 3]",
		},
		{
			Ints([]float64{2}),
			"[2]",
		},
		{
			Ints(-1.0),
			"[-1]",
		},
		{
			Ints(math.NaN()),
			"[NaN]",
		},
		{
			Ints(math.Inf(1)),
			"[NaN]",
		},
		{
			Ints(math.Inf(-1)),
			"[NaN]",
		},
		{
			Ints([]bool{true, true, false}),
			"[1 1 0]",
		},
		{
			Ints([]bool{false}),
			"[0]",
		},
		{
			Ints(true),
			"[1]",
		},
		{
			Ints([]int{}),
			"[]",
		},
		{
			Ints(nil),
			"[]",
		},
		{
			Ints(Strings([]string{"1", "2", "3"})),
			"[1 2 3]",
		},
		{
			Ints(Ints([]string{"1", "2", "3"})),
			"[1 2 3]",
		},
	}
	for testnum, test := range table {
		if err := test.series.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(test.series)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(test.series); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
	}
}

func TestUints(t *testing.T) {
	table := []struct {
		series   Series
		expected string
	}{
		{
			Uints([]string{"A", "B", "1", "2"}),
			"[NaN NaN 1 2]",
		},
		{
			Uints([]string{"A", "", "1", "2"}),
			"[NaN  1 2]",
		},
		{
			New(Uints([]string{"A", "", "1", "2"}), Uint, ""),
			"[NaN  1 2]",
		},
		{
			Uints([]string{"1"}),
			"[1]",
		},
		{
			Uints("2"),
			"[2]",
		},
		{
			Uints([]int{1, 2, 3}),
			"[1 2 3]",
		},
		{
			Uints([]int{2}),
			"[2]",
		},
		{
			Uints(-1),
			"[18446744073709551615]",
		},
		{
			Uints([]float64{1, 2, 3}),
			"[1 2 3]",
		},
		{
			Uints([]float64{2}),
			"[2]",
		},
		{
			Uints(-1.0),
			"[18446744073709551615]",
		},
		{
			Uints(math.NaN()),
			"[NaN]",
		},
		{
			Uints(math.Inf(1)),
			"[NaN]",
		},
		{
			Uints(math.Inf(-1)),
			"[NaN]",
		},
		{
			Uints([]bool{true, true, false}),
			"[1 1 0]",
		},
		{
			Uints([]bool{false}),
			"[0]",
		},
		{
			Uints(true),
			"[1]",
		},
		{
			Uints([]int{}),
			"[]",
		},
		{
			Uints(nil),
			"[]",
		},
		{
			Uints(Strings([]string{"1", "2", "3"})),
			"[1 2 3]",
		},
		{
			Uints(Uints([]string{"1", "2", "3"})),
			"[1 2 3]",
		},
	}
	for testnum, test := range table {
		if err := test.series.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(test.series)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(test.series); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
	}
}

func TestFloats(t *testing.T) {
	table := []struct {
		series   Series
		expected string
	}{
		{
			Floats([]string{"A", "B", "1", "2"}),
			"[NaN NaN 1.000000 2.000000]",
		},
		{
			Floats([]string{"NaN", "", "1", "2"}),
			"[NaN  1.000000 2.000000]",
		},
		{
			New(Floats([]string{"NaN", "", "1", "2"}), Float, ""),
			"[NaN  1.000000 2.000000]",
		},
		{
			New(Ints([]string{"NaN", "", "1", "2"}), Float, ""),
			"[NaN  1.000000 2.000000]",
		},
		{
			Floats([]string{"1"}),
			"[1.000000]",
		},
		{
			Floats("2.1"),
			"[2.100000]",
		},
		{
			Floats([]int{1, 2, 3}),
			"[1.000000 2.000000 3.000000]",
		},
		{
			Floats([]int{2}),
			"[2.000000]",
		},
		{
			Floats(-1),
			"[-1.000000]",
		},
		{
			Floats([]float64{1.1, 2, 3}),
			"[1.100000 2.000000 3.000000]",
		},
		{
			Floats([]float64{2}),
			"[2.000000]",
		},
		{
			Floats(-1.0),
			"[-1.000000]",
		},
		{
			Floats(math.NaN()),
			"[NaN]",
		},
		{
			Floats(math.Inf(1)),
			"[+Inf]",
		},
		{
			Floats(math.Inf(-1)),
			"[-Inf]",
		},
		{
			Floats([]bool{true, true, false}),
			"[1.000000 1.000000 0.000000]",
		},
		{
			Floats([]bool{false}),
			"[0.000000]",
		},
		{
			Floats(true),
			"[1.000000]",
		},
		{
			Floats([]int{}),
			"[]",
		},
		{
			Floats(nil),
			"[]",
		},
		{
			Floats(Strings([]string{"1", "2", "3"})),
			"[1.000000 2.000000 3.000000]",
		},
	}
	for testnum, test := range table {
		if err := test.series.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(test.series)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(test.series); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
	}
}

func TestBools(t *testing.T) {
	table := []struct {
		series   Series
		expected string
	}{
		{
			Bools([]string{"A", "true", "1", "f"}),
			"[ true true false]",
		},
		{
			Bools([]string{"t"}),
			"[true]",
		},
		{
			Bools("False"),
			"[false]",
		},
		{
			Bools([]int{1, 2, 0}),
			"[true true false]",
		},
		{
			Bools([]int{1}),
			"[true]",
		},
		{
			Bools(-1),
			"[true]",
		},
		{
			Bools([]float64{1, 2, 0}),
			"[true true false]",
		},
		{
			Bools([]float64{0}),
			"[false]",
		},
		{
			Bools(-1.0),
			"[true]",
		},
		{
			Bools(math.NaN()),
			"[false]",
		},
		{
			Bools(math.Inf(1)),
			"[true]",
		},
		{
			Bools(math.Inf(-1)),
			"[true]",
		},
		{
			Bools([]bool{true, true, false}),
			"[true true false]",
		},
		{
			Bools([]bool{false}),
			"[false]",
		},
		{
			Bools(true),
			"[true]",
		},
		{
			Bools([]int{}),
			"[]",
		},
		{
			Bools(nil),
			"[]",
		},
		{
			Bools(Strings([]string{"1", "0", "1"})),
			"[true false true]",
		},
	}
	for testnum, test := range table {
		if err := test.series.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		received := fmt.Sprint(test.series)
		if expected != received {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		if err := checkTypes(test.series); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
	}
}

// TODO: finish adding tests
func TestSeries_Factorize(t *testing.T) {
	table := []struct {
		series  Series
		codes   Series
		unqiues Series
		sort    bool
	}{
		{
			Strings([]string{"b", "b", "a", "c", "b"}),
			Ints([]int{0, 0, 1, 2, 0}),
			Strings([]string{"b", "a", "c"}),
			false,
		},
		{
			Strings([]string{"b", "b", "a", "c", "b"}),
			Ints([]int{1, 1, 0, 2, 1}),
			Strings([]string{"a", "b", "c"}),
			true,
		},
		{
			Strings([]string{"b", "b", "a", "c", ""}),
			Ints([]int{0, 0, 1, 2, -1}),
			Strings([]string{"b", "a", "c"}),
			false,
		},
		{
			Floats([]float64{3, 1.1, 2, 3}),
			Ints([]int{0, 1, 2, 0}),
			Floats([]float64{3, 1.1, 2}),
			false,
		},
		{
			Floats([]float64{3, 1.1, 2, 3}),
			Ints([]int{2, 0, 1, 2}),
			Floats([]float64{1.1, 2, 3}),
			true,
		},
	}
	for testnum, test := range table {
		b := test.series
		bc, bu := b.Factorize(test.sort)
		if bc.Err != nil {
			t.Errorf("Test %d: %v\n", testnum, bc.Err)
			continue
		}
		if b.Type() != bu.Type() {
			t.Errorf("Test %d Mismatched types:\n\tExpected: %v\n\tReceived: %v\n", testnum, b.Type(), bu.Type())
			continue
		}
		expectedCodes, _ := test.codes.Records(false)
		receivedCodes, _ := bc.Records(false)
		if !reflect.DeepEqual(expectedCodes, receivedCodes) {
			t.Errorf("Test %d Codes:\n\tExpected: %v\n\tReceived: %v\n", testnum, expectedCodes, receivedCodes)
		}
		expectedUniques, _ := test.unqiues.Records(false)
		receivedUniques, _ := bu.Records(false)
		if !reflect.DeepEqual(expectedUniques, receivedUniques) {
			t.Errorf("Test %d Uniques:\n\tExpected: %v\n\tReceived: %v\n", testnum, expectedUniques, receivedUniques)
		}
	}
}

func TestSeries_Copy(t *testing.T) {
	tests := []Series{
		Strings([]string{"1", "2", "3", "a", "b", "c"}),
		Ints([]string{"1", "2", "3", "a", "b", "c"}),
		Floats([]string{"1", "2", "3", "a", "b", "c"}),
		Bools([]string{"1", "0", "1", "t", "f", "c"}),
	}
	for testnum, test := range tests {
		a := test
		b := a.Copy()
		if fmt.Sprint(a) != fmt.Sprint(b) {
			t.Error("Different values when copying String elements")
		}
		if err := b.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		if err := checkTypes(b); err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		//if err := checkAddr(a.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nA:%v\nB:%v", testnum, err, a.Addr(), b.Addr())
		//}
	}
}

func TestSeries_Records(t *testing.T) {
	tests := []struct {
		series   Series
		expected []string
	}{
		{
			Strings([]string{"1", "2", "3", "a", "b", "c"}),
			[]string{"1", "2", "3", "a", "b", "c"},
		},
		{
			Ints([]string{"1", "2", "3", "a", "b", "c"}),
			[]string{"1", "2", "3", "NaN", "NaN", "NaN"},
		},
		{
			Ints([]string{"1", "2", "3", "a", "b", ""}),
			[]string{"1", "2", "3", "NaN", "NaN", ""},
		},
		{
			Uints([]string{"1", "2", "3", "a", "b", "c"}),
			[]string{"1", "2", "3", "NaN", "NaN", "NaN"},
		},
		{
			Uints([]string{"1", "2", "3", "a", "b", ""}),
			[]string{"1", "2", "3", "NaN", "NaN", ""},
		},
		{
			Floats([]string{"1", "2", "3", "a", "b", "c"}),
			[]string{"1.000000", "2.000000", "3.000000", "NaN", "NaN", "NaN"},
		},
		{
			Floats([]string{"1", "2", "3", "a", "b", ""}),
			[]string{"1.000000", "2.000000", "3.000000", "NaN", "NaN", ""},
		},
		{
			Bools([]string{"1", "0", "1", "t", "f", "c"}),
			[]string{"true", "false", "true", "true", "false", ""},
		},
	}
	for testnum, test := range tests {
		expected := test.expected
		received, _ := test.series.Records(true) // force convertions
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Float(t *testing.T) {
	precision := 0.0000001
	floatEquals := func(x, y []float64) bool {
		if len(x) != len(y) {
			return false
		}
		for i := 0; i < len(x); i++ {
			a := x[i]
			b := y[i]
			if (a-b) > precision || (b-a) > precision {
				return false
			}
		}
		return true
	}
	tests := []struct {
		series   Series
		expected []float64
	}{
		{
			Strings([]string{"1", "2", "3", "a", "b", "c"}),
			[]float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN()},
		},
		{
			Ints([]string{"1", "2", "3", "a", "b", "c"}),
			[]float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN()},
		},
		{
			Floats([]string{"1", "2", "3", "a", "b", "c"}),
			[]float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN()},
		},
		{
			Bools([]string{"1", "0", "1", "t", "f", "c"}),
			[]float64{1, 0, 1, 1, 0, math.NaN()},
		},
	}
	for testnum, test := range tests {
		expected := test.expected
		received, _ := test.series.Float(true)
		if !floatEquals(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_New(t *testing.T) {
	//b := Ints([]string{"1", "2", "3"})
	a := Ints([]string{"a", "4", "c"})
	na := New(a, a.t, a.Name)
	received := fmt.Sprint(na)
	if received != "[NaN 4 NaN]" {
		t.Errorf("Expected:\n[NaN 4 NaN]\nReceived:\n%v", received)
	}
}

func TestSeries_Concat(t *testing.T) {
	tests := []struct {
		a        Series
		b        Series
		expected []string
	}{
		{
			Strings([]string{"1", "2", "3"}),
			Strings([]string{"a", "b", "c"}),
			[]string{"1", "2", "3", "a", "b", "c"},
		},
		{
			Ints([]string{"1", "2", "3"}),
			Ints([]string{"a", "4", "c"}),
			[]string{"1", "2", "3", "NaN", "4", "NaN"},
		},
		{
			Uints([]string{"1", "2", "3"}),
			Uints([]string{"a", "4", "c"}),
			[]string{"1", "2", "3", "NaN", "4", "NaN"},
		},
		{
			Floats([]string{"1", "2", "3"}),
			Floats([]string{"a", "4", "c"}),
			[]string{"1.000000", "2.000000", "3.000000", "NaN", "4.000000", "NaN"},
		},
		{
			Bools([]string{"1", "1", "0"}),
			Bools([]string{"0", "0", "0"}),
			[]string{"true", "true", "false", "false", "false", "false"},
		},
	}
	for testnum, test := range tests {
		ab := test.a.Concat(test.b)
		if err := ab.Err; err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		received, _ := ab.Records(false)
		expected := test.expected
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
		//a := test.a
		//b := ab
		//if err := checkAddr(a.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nA:%v\nAB:%v", testnum, err, a.Addr(), b.Addr())
		//}
		//a = test.b
		//b = ab
		//if err := checkAddr(a.Addr(), b.Addr()); err != nil {
		//t.Errorf("Test:%v\nError:%v\nB:%v\nAB:%v", testnum, err, a.Addr(), b.Addr())
		//}
	}
}

func TestSeries_Order(t *testing.T) {
	tests := []struct {
		series   Series
		reverse  bool
		expected []int
	}{
		{
			Ints([]string{"2", "1", "3", "NaN", "4", ""}),
			false,
			[]int{1, 0, 2, 4, 3, 5}, // 1,2,3,4,NaN,<nil>
		},
		{
			Ints([]string{"2", "1", "3", "", "4", "NaN"}),
			false,
			[]int{1, 0, 2, 4, 3, 5},
		},
		{
			Uints([]string{"2", "1", "3", "NaN", "4", ""}),
			false,
			[]int{1, 0, 2, 4, 3, 5},
		},
		{
			Uints([]string{"2", "1", "3", "", "4", "NaN"}),
			false,
			[]int{1, 0, 2, 4, 3, 5},
		},
		{
			Floats([]string{"2", "1", "3", "NaN", "4", "NaN"}),
			false,
			[]int{1, 0, 2, 4, 3, 5},
		},
		{
			Strings([]string{"b", "a", "c"}),
			false,
			[]int{1, 0, 2},
		},
		{
			Bools([]bool{true, false, false, false, true}),
			false,
			[]int{1, 2, 3, 0, 4},
		},
		{
			Ints([]string{"2", "1", "3", "NaN", "4", ""}),
			true,
			[]int{4, 2, 0, 1, 3, 5},
		},
		{
			Floats([]string{"2", "1", "3", "NaN", "4", "NaN"}),
			true,
			[]int{4, 2, 0, 1, 3, 5},
		},
		{
			Strings([]string{"b", "c", "a"}),
			true,
			[]int{1, 0, 2},
		},
		{
			Bools([]bool{true, false, false, false, true}),
			true,
			[]int{0, 4, 1, 2, 3},
		},
	}
	for testnum, test := range tests {
		received := test.series.Order(test.reverse)
		expected := test.expected
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_IsNaN(t *testing.T) {
	tests := []struct {
		series   Series
		expected []bool
	}{
		{
			Ints([]string{"2", "1", "3", "NaN", "4", "NaN"}),
			[]bool{false, false, false, true, false, true},
		},
		{
			Uints([]string{"2", "1", "3", "NaN", "4", "NaN"}),
			[]bool{false, false, false, true, false, true},
		},
		{
			Floats([]string{"A", "1", "B", "3"}),
			[]bool{true, false, true, false},
		},
		{
			Floats([]string{"A", "1", "", "3"}),
			[]bool{true, false, true, false},
		},
		{
			Bools([]string{"1", "0", "A"}),
			[]bool{false, false, true},
		},
		{
			Bools([]string{"1", "0", ""}),
			[]bool{false, false, true},
		},
	}
	for testnum, test := range tests {
		received := test.series.IsNaN()
		expected := test.expected
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_IsValid(t *testing.T) {
	tests := []struct {
		series   Series
		expected []bool
	}{
		{
			Ints([]string{"2", "1", "3", "", "4", "A"}),
			[]bool{true, true, true, false, true, true},
		},
		{
			Uints([]string{"2", "1", "3", "", "4", ""}),
			[]bool{true, true, true, false, true, false},
		},
		{
			Floats([]string{"", "1", "A", "3"}),
			[]bool{false, true, true, true},
		},
		{
			Strings([]string{"", "1", "A", "3"}),
			[]bool{false, true, true, true},
		},
		{
			Bools([]string{"1", "0", "A"}),
			[]bool{true, true, false},
		},
		{
			Bools([]string{"1", "0", ""}),
			[]bool{true, true, false},
		},
	}
	for testnum, test := range tests {
		received := test.series.IsValid()
		expected := test.expected
		if !reflect.DeepEqual(expected, received) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_StdDev(t *testing.T) {
	tests := []struct {
		series   Series
		expected float64
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			3.02765,
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			1.0,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			math.NaN(),
		},
		{
			Bools([]bool{true, true, false, true}),
			0.5,
		},
		{
			Floats([]float64{}),
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.StdDev()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Mean(t *testing.T) {
	tests := []struct {
		series   Series
		expected float64
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			5.5,
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			2.0,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			math.NaN(),
		},
		{
			Bools([]bool{true, true, false, true}),
			0.75,
		},
		{
			Floats([]float64{}),
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.Mean()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Max(t *testing.T) {
	tests := []struct {
		series   Series
		expected float64
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			10,
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			3.0,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			math.NaN(),
		},
		{
			Bools([]bool{true, true, false, true}),
			1.0,
		},
		{
			Floats([]float64{}),
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.Max()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Median(t *testing.T) {
	tests := []struct {
		series   Series
		expected float64
	}{
		{
			// Extreme observations should not factor in.
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000}),
			7,
		},
		{
			// Change in order should influence result.
			Ints([]int{1, 2, 3, 10, 100, 1000, 10000, 4, 5, 6, 7, 8, 9}),
			7,
		},
		{
			Floats([]float64{20.2755, 4.98964, -20.2006, 1.19854, 1.89977,
				1.51178, -17.4687, 4.65567, -8.65952, 6.31649,
			}),
			1.705775,
		},
		{
			// Change in order should not influence result.
			Floats([]float64{4.98964, -20.2006, 1.89977, 1.19854,
				1.51178, -17.4687, -8.65952, 20.2755, 4.65567, 6.31649,
			}),
			1.705775,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			math.NaN(),
		},
		{
			Bools([]bool{true, true, false, true}),
			math.NaN(),
		},
		{
			Floats([]float64{}),
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.Median()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Min(t *testing.T) {
	tests := []struct {
		series   Series
		expected float64
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			1.0,
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			1.0,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			math.NaN(),
		},
		{
			Bools([]bool{true, true, false, true}),
			0.0,
		},
		{
			Floats([]float64{}),
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.Min()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_MaxStr(t *testing.T) {
	tests := []struct {
		series   Series
		expected string
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			"",
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			"",
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			"D",
		},
		{
			Strings([]string{"quick", "Brown", "fox", "Lazy", "dog"}),
			"quick",
		},
		{
			Bools([]bool{true, true, false, true}),
			"",
		},
		{
			Floats([]float64{}),
			"",
		},
	}

	for testnum, test := range tests {
		received, err := test.series.MaxStr()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if received != expected {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_MinStr(t *testing.T) {
	tests := []struct {
		series   Series
		expected string
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			"",
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			"",
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			"A",
		},
		{
			Strings([]string{"quick", "Brown", "fox", "Lazy", "dog"}),
			"Brown",
		},
		{
			Bools([]bool{true, true, false, true}),
			"",
		},
		{
			Floats([]float64{}),
			"",
		},
	}

	for testnum, test := range tests {
		received, err := test.series.MinStr()
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if received != expected {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Quantile(t *testing.T) {
	tests := []struct {
		series   Series
		p        float64
		expected float64
	}{
		{
			Ints([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			0.9,
			9,
		},
		{
			Floats([]float64{3.141592, math.Sqrt(3), 2.718281, math.Sqrt(2)}),
			0.8,
			3.141592,
		},
		{
			Floats([]float64{1.0, 2.0, 3.0}),
			0.5,
			2.0,
		},
		{
			Strings([]string{"A", "B", "C", "D"}),
			0.25,
			math.NaN(),
		},
		{
			Bools([]bool{false, false, false, true}),
			0.75,
			0.0,
		},
		{
			Floats([]float64{}),
			0.50,
			math.NaN(),
		},
	}

	for testnum, test := range tests {
		received, err := test.series.Quantile(test.p)
		if err != nil {
			t.Errorf("Test:%v\nError:%v", testnum, err)
		}
		expected := test.expected
		if !compareFloats(received, expected, 6) {
			t.Errorf(
				"Test:%v\nExpected:\n%v\nReceived:\n%v",
				testnum, expected, received,
			)
		}
	}
}

func TestSeries_Map(t *testing.T) {
	tests := []struct {
		series   Series
		expected Series
	}{
		{
			Bools([]bool{false, true, false, false, true}),
			Bools([]bool{false, true, false, false, true}),
		},
		{
			Floats([]float64{1.5, -3.23, -0.337397, -0.380079, 1.60979, 34.}),
			Floats([]float64{3, -6.46, -0.674794, -0.760158, 3.21958, 68.}),
		},
		{
			Floats([]float64{math.Pi, math.Phi, math.SqrtE, math.Cbrt(64)}),
			Floats([]float64{2 * math.Pi, 2 * math.Phi, 2 * math.SqrtE, 2 * math.Cbrt(64)}),
		},
		{
			Strings([]string{"XyZApple", "XyZBanana", "XyZCitrus", "XyZDragonfruit"}),
			Strings([]string{"Apple", "Banana", "Citrus", "Dragonfruit"}),
		},
		{
			Strings([]string{"San Francisco", "XyZTokyo", "MoscowXyZ", "XyzSydney"}),
			Strings([]string{"San Francisco", "Tokyo", "MoscowXyZ", "XyzSydney"}),
		},
		{
			Ints([]int{23, 13, 101, -64, -3}),
			Ints([]int{28, 18, 106, -59, 2}),
		},
		{
			Ints([]string{"morning", "noon", "afternoon", "evening", "night"}),
			Ints([]int{5, 5, 5, 5, 5}),
		},
	}

	doubleFloat64 := func(e Element) Element {
		var result Element
		result = e.Copy()
		f, err := result.Float()
		if err != nil {
			t.Errorf("%v", err)
			return Element(nil)
		}
		result.Set(f * 2)
		return Element(result)
	}

	// and two booleans
	and := func(e Element) Element {
		var result Element
		result = e.Copy()
		b, err := result.Bool()
		if err != nil {
			t.Errorf("%v", err)
			return Element(nil)
		}
		result.Set(b && true)
		return Element(result)
	}

	// add constant (+5) to value (v)
	add5Int := func(e Element) Element {
		var result Element
		result = e.Copy()
		i, err := result.Int()
		if err != nil {
			return Element(&IntElement{
				e:     +5,
				valid: true,
			})
		}
		result.Set(i + 5)
		return Element(result)
	}

	// trim (XyZ) prefix from string
	trimXyZPrefix := func(e Element) Element {
		var result Element
		result = e.Copy()
		s, err := result.String()
		if err != nil {
			t.Errorf("%v", err)
			return Element(nil)
		}
		result.Set(strings.TrimPrefix(s, "XyZ"))
		return Element(result)
	}

	for testnum, test := range tests {
		switch test.series.Type() {
		case Bool:
			expected := test.expected
			received := test.series.Map(and)
			for i := 0; i < expected.Len(); i++ {
				e, _ := expected.Elem(i).Bool()
				r, _ := received.Elem(i).Bool()

				if e != r {
					t.Errorf(
						"Test:%v\nExpected:\n%v\nReceived:\n%v",
						testnum, expected, received,
					)
				}
			}

		case Float:
			expected := test.expected
			received := test.series.Map(doubleFloat64)
			for i := 0; i < expected.Len(); i++ {
				e, _ := expected.Elem(i).Float()
				r, _ := received.Elem(i).Float()
				if !compareFloats(e, r, 6) {
					t.Errorf(
						"Test:%v\nExpected:\n%v\nReceived:\n%v",
						testnum, expected, received,
					)
				}
			}
		case Int:
			expected := test.expected
			received := test.series.Map(add5Int)
			for i := 0; i < expected.Len(); i++ {
				e, _ := expected.Elem(i).Int()
				r, _ := received.Elem(i).Int()
				if e != r {
					t.Errorf(
						"Test:%v\nExpected:\n%v\nReceived:\n%v",
						testnum, expected, received,
					)
				}
			}
		case String:
			expected := test.expected
			received := test.series.Map(trimXyZPrefix)
			for i := 0; i < expected.Len(); i++ {
				e, _ := expected.Elem(i).String()
				r, _ := received.Elem(i).String()
				if strings.Compare(e, r) != 0 {
					t.Errorf(
						"Test:%v\nExpected:\n%v\nReceived:\n%v",
						testnum, expected, received,
					)
				}
			}
		default:
		}
	}
}
