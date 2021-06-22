# Series

This is forked from [gota](https://github.com/go-gota/gota) with the aim of being a bridge between [SciDB](https://www.paradigm4.com/) and the Apache [Arrow format](https://arrow.apache.org/). For this reason, this version makes a distinction between missing values (`NA`, `nil`, `null`, `None`) and "not a number" values (`NaNs`). The latter, being defined by [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754) only for floating point numbers. The original package seems to follow what Python Pandas does, see [NaN, Integer NA values and NA type promotions](https://pandas.pydata.org/pandas-docs/version/1.1.0/user_guide/gotchas.html#nan-integer-na-values-and-na-type-promotions). In both cases, `NAs` are promoted to `NaN` in various ways, thus **creating** values that did not exist. 

## SciDB Types

Internally, Float, Int and Uint types are backed by 64-bits.

| SciDB Data Type | Series Type |
|:----------------|:------------|
| bool              |   Bool    |
| double (float64)  |   Float   |
| float (float32)   |   Float   |
| int8 to int64     |   Int     |
| uint8 to uint64   |   Uint    |
| string            |   String  |
| char              |   n.s.    |
| date              |   n.s.    |
| time              |   n.s.    |
| timestamp         |   n.s.    |
| datetime          |   n.s.    |
| datetimez         |   n.s.    |
| interval          |    n.s.    |


## Promotions

All series types have a flag for NA that is backed by a boolean. 

### Promotions to NaN

Series uses a special value `NaNElement` for non-floats internally; floats are IEEE 754, so they have a value for `NaN`. This is preserved when outputting any `NaN` as a string (*"NaN"*) or a float (as the native NaN representation). Converting to other types results in an error and the actual value is unspecified.

Using `NaN` as a string and converting to the other types results in the NaN flag being set and the element returning `NaNElement` or `isNaN` as **true**.

When converting other types to a float, any non-NA value that does not evaluate to a float returns an NaN.

### Promotions to NA

Creating/setting a value to `nil` sets the NA flag to `true`. Another way to create an NA is using an empty string ``.

The native value stored is undefined and should **NOT** be trusted when `isValid` returns **false**. That is all of the accessor/conversion methods will return an error. However, to prevent panics (crashing out), these are the values returned:

| type | value |
|:----:|:-----:|
| float | NaN |
| int | 0 |
| uint | 0 |
| boolean | 0/false |
| string | `` |

## Comparisons

### Equals

Every NaN shall compare unordered with everything, including itself. However, this isn't the same for nils.

| A | B | comparison | result |
|:-:|:-:|:----------:|:------:|
| NaN | NaN | Eq | false |
| NaN | NaN | Neq | **true** |
| NaN | value | Eq | false |
| NaN | value | Neq | false |
| nil | nil | Eq | **true** |
| nil | value | Eq | false |
| nil | nil | Neq | false |
| nil | value | Neq| **true** |
| nil | NaN | Eq | false |
| nil | NaN | Neq | **true** |

Note, IEEE 754 says that only NaNs satisfy `f != f`!

### Less/Greater-thans

Every NaN and nil shall compare unordered with everything, including itself. For example:

| A | B | comparison | result |
|:-:|:-:|:----------:|:------:|
| NaN | NaN | Less | false |
| NaN | NaN | LessEq | false |
| NaN | value | Less | false |
| NaN | value | LessEq | false |
| nil | nil | Less | false |
| nil | nil | LessEq | false |
| nil | value | Less | false |
| nil | value | LessEq | false |

The table is the same using **Greater** or **GreaterEq**.

### In Comparison

A big gotcha is that `NaN` can't be tested using the **In** comparator because of the rules above. One must use `HasNaN()` or `IsNaN()`.

## Examples

### NA vs NaN

#### Starting with Ints

```go
a := Ints([]string{"", "2", "1", "5", "NaN"})
a.Elem(0).isValid() // returns false
a.Elem(0).isNaN()   // returns true
a.Elem(4).isValid() // returns true
a.Elem(4).isNaN()   // returns true
```

Forcing promotion of nils to NaN:

```go
// Float returns the elements of a Series as a []float64.
b, err := a.Float(false) // err is non-nil and the state of b is unknown
c, err := a.Float(true)  // err should be nil and c = {NaN, 2.0, 1.0, 5.0, NaN}
c.Elem(0).isValid()      // returns true now
d, err := a.Elem(0).Float() // err is non-nil and d == NaN
```

Converting to a float series:

```go
e := Floats(a)
e.Elem(0).isValid() // returns false
e.Elem(0).isNaN()   // returns true
e.Elem(4).isValid() // returns true
e.Elem(4).isNaN()   // returns true
```

#### Starting with Floats

```go
a := Floats([]string{"", "2", "1", "5", "NaN"})
a.Elem(0).isValid() // returns false
a.Elem(0).isNaN()   // returns true
a.Elem(4).isValid() // returns true
a.Elem(4).isNaN()   // returns true
```

Forcing promotion of nils to NaN:

```go
b, err := a.Float(false) // err is non-nil and the state of b is unknown
c, err := a.Float(true)  // err should be nil and c = {NaN, 2.0, 1.0, 5.0, NaN}
c.Elem(0).isValid()      // returns true now
```

## Random

### Floating Point Math

So, not all NaNs are created equal. There are issues converting float32 NaN to float64 NaN in golang; see [Issue 36399](https://github.com/golang/go/issues/36399).
It seems the safe way of testing is using `math.IsNaN(f)` which is `return f != f` and IEEE 754 says that only NaNs satisfy this. The following doesn't work (or did not):

```go
package main

import (
    "fmt" 
    "math"
)

func main() {
    a := float32(math.NaN())
    fmt.Println(a) // prints out "NaN"
    if float64(a) == math.NaN() { // is false
        fmt.Println("math.NaN")  
    }
    if math.IsNaN(float64(a)) { // is true
        fmt.Println("math.IsNaN")
    }
    if a != a { // is true
        fmt.Println("a != a")
    }
    // float32 sNaN (uvnan); uint32(0x7F800001)
    // float64 sNaN (uvnan); unit64(0x7FF8000000000001)
}
```
