// Paradigm4

package dataframe

// TODO: what is the difference between Valid and Null? If isValid then !isNull and !isValid then isNull

import (
	"fmt"

	"github.com/Paradigm4/gota/series"
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/float16"
	"github.com/apache/arrow/go/arrow/memory"
)

// Arrow Array Table to a DataFrame
func TableToDataframe(tbl array.Table) DataFrame {
	columns := make([]series.Series, tbl.NumCols())
	for i := 0; i < int(tbl.NumCols()); i++ {
		col, err := arrayColumnToSeries(tbl.Column(i))
		if err != nil {
			return DataFrame{Err: err}
		}
		columns[i] = col
	}
	return New(columns...)
}

func arrayColumnToSeries(column *array.Column) (series.Series, error) {

	var s series.Series
	switch column.DataType().ID() {
	case arrow.BOOL:
		s = series.New(nil, series.Bool, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewBooleanData(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.INT8:
		s = series.New(nil, series.Int, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewInt8Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.INT16:
		s = series.New(nil, series.Int, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewInt16Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.INT32:
		s = series.New(nil, series.Int, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewInt32Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.INT64:
		s = series.New(nil, series.Int, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewInt64Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.UINT8:
		s = series.New(nil, series.Uint, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewUint8Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.UINT16:
		s = series.New(nil, series.Uint, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewUint16Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Elem(i).Set(data.Value(j))
				}
				i++
			}
		}
	case arrow.UINT32:
		s = series.New(nil, series.Uint, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewUint32Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.UINT64:
		s = series.New(nil, series.Uint, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewUint64Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.FLOAT16:
		s = series.New(nil, series.Float, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewFloat16Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j).Float32())
				}
				i++
			}
		}
	case arrow.FLOAT32:
		s = series.New(nil, series.Float, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewFloat32Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.FLOAT64:
		s = series.New(nil, series.Float, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewFloat64Data(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	case arrow.STRING:
		s = series.New(nil, series.String, column.Name(), column.Len())
		i := 0
		for _, c := range column.Data().Chunks() {
			data := array.NewStringData(c.Data())
			for j := 0; j < data.Len(); j++ {
				if data.IsValid(j) && !data.IsNull(j) {
					s.Set(i, data.Value(j))
				}
				i++
			}
		}
	default:
		return series.Series{}, fmt.Errorf("unsupported Arrow Type: %v", column.DataType())
	}
	return s, nil
}

// Convert a DataFrame to an Arrow Table
// Order of columns (by name) is given by colNames. If empty, this uses the existing order of columns in the DataFrame
/*
func DataframeToTable(df DataFrame, colNames ...string) (array.Table, error) {
*/

// Caller is responsible for Releasing the record
// Arrays are arranged in the order given in the schema (fields)
func DataframeToRecordWithSchema(df DataFrame, schema *arrow.Schema, mem memory.Allocator) (array.Record, error) {
	if schema.Fields() == nil || len(schema.Fields()) == 0 {
		return nil, fmt.Errorf("[DataframeToRecordWithSchema] No fields/columns given")
	}

	for _, f := range schema.Fields() {
		s := df.Col(f.Name)
		if s.Err != nil {
			return nil, s.Err
		}
	}
	rb := array.NewRecordBuilder(mem, schema)
	defer rb.Release()

	for f, ab := range rb.Fields() {
		s := df.Col(schema.Field(f).Name)
		l := s.Len()
		ab.Reserve(l) // pre-allocate space so we can call UnsafeAppends (all expect for String which is lacking it)
		switch schema.Field(f).Type.ID() {
		case arrow.BOOL:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Bool()
					if err != nil {
						return nil, err
					}
					ab.(*array.BooleanBuilder).UnsafeAppend(v)
				} else {
					ab.(*array.BooleanBuilder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.INT8:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Int()
					if err != nil {
						return nil, err
					}
					ab.(*array.Int8Builder).UnsafeAppend(int8(v))
				} else {
					ab.(*array.Int8Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.INT16:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Int()
					if err != nil {
						return nil, err
					}
					ab.(*array.Int16Builder).UnsafeAppend(int16(v))
				} else {
					ab.(*array.Int16Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.INT32:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Int()
					if err != nil {
						return nil, err
					}
					ab.(*array.Int32Builder).UnsafeAppend(int32(v))
				} else {
					ab.(*array.Int32Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.INT64:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Int()
					if err != nil {
						return nil, err
					}
					ab.(*array.Int64Builder).UnsafeAppend(int64(v))
				} else {
					ab.(*array.Int64Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.UINT8:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Uint()
					if err != nil {
						return nil, err
					}
					ab.(*array.Uint8Builder).UnsafeAppend(uint8(v))
				} else {
					ab.(*array.Uint8Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.UINT16:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Uint()
					if err != nil {
						return nil, err
					}
					ab.(*array.Uint16Builder).UnsafeAppend(uint16(v))
				} else {
					ab.(*array.Uint16Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.UINT32:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Uint()
					if err != nil {
						return nil, err
					}
					ab.(*array.Uint32Builder).UnsafeAppend(uint32(v))
				} else {
					ab.(*array.Uint32Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.UINT64:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Uint()
					if err != nil {
						return nil, err
					}
					ab.(*array.Uint64Builder).UnsafeAppend(v)
				} else {
					ab.(*array.Uint64Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.FLOAT16:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Float()
					if err != nil {
						return nil, err
					}
					ab.(*array.Float16Builder).UnsafeAppend(float16.New(float32(v)))
				} else {
					ab.(*array.Float16Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.FLOAT32:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Float()
					if err != nil {
						return nil, err
					}
					ab.(*array.Float32Builder).UnsafeAppend(float32(v))
				} else {
					ab.(*array.Float32Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.FLOAT64:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.Float()
					if err != nil {
						return nil, err
					}
					ab.(*array.Float64Builder).UnsafeAppend(v)
				} else {
					ab.(*array.Float64Builder).UnsafeAppendBoolToBitmap(false)
				}
			}
		case arrow.STRING:
			for i := 0; i < l; i++ {
				e := s.Elem(i)
				if e.IsValid() {
					v, err := e.String()
					if err != nil {
						return nil, err
					}
					ab.(*array.StringBuilder).Append(v)
				} else {
					ab.(*array.StringBuilder).AppendNull()
				}
			}
		}
	}

	r := rb.NewRecord()
	// defer r.Release() don't release the caller needs to do this

	return r, nil
}
