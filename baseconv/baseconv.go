// Package baseconv converts numeric strings between arbitrary radices
//
// The math is done manually, so values are not restricted by system maximums
package baseconv

import (
	"fmt"
	"strconv"
)

var ErrDuplicateCharacter = fmt.Errorf("duplicate character")

const chars64 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	Base2,  _ = NewBaseMap(chars64[:2])
	Base8,  _ = NewBaseMap(chars64[:8])
	Base10, _ = NewBaseMap(chars64[:10])
	Base16, _ = NewBaseMap(chars64[:16])
	Base36, _ = NewBaseMap(chars64[:36])
	Base62, _ = NewBaseMap(chars64)
)

// BaseMap represents a pool of characters used for a certain radix
type BaseMap struct {
	radix  int
	values map[rune]int
	chars  map[int]rune
}

// Parse converts a numeric string to a Number
func (bm *BaseMap) Parse(input string) (*Number, error) {
	// Use len(input) even though it returns the number of bytes not runes.
	// Some memory might be wasted, but it's way faster than utf8.RuneCountInString (over 100x faster)
	values := make([]int, 0, len(input))

	// add the value of each digit to values
	for _, c := range input {
		value, ok := bm.values[c]

		// gotta check for that character validity ;)
		if !ok {
			return nil, fmt.Errorf("invalid character %c", c)
		}

		values = append(values, value)
	}

	return &Number{vals: values, bmap: bm}, nil
}

// Number represents a number with an arbitrary radix
type Number struct {
	vals []int
	bmap *BaseMap
}

// Int64 returns the int64 value of the number as long as it is within 64 bits
func (num *Number) Int64() int64 {
	intVal, _ := strconv.ParseInt(num.Format(Base10), 10, 64)
	return intVal
}

// Format converts the number to a numeric string according to the given basemap
func (num *Number) Format(bm *BaseMap) string {
	radix, divisor := num.bmap.radix, bm.radix

	values := num.vals[:]
	places := len(values)
	str := make([]rune, 0, places)
	for places > 0 {
		// do manual division
		var remainder, place int
		for i := 0; i < places; i++ {
			dividend := remainder * radix + values[i]
			if dividend >= divisor {
				remainder = dividend % divisor
			} else if i == 0 {
				remainder = dividend
				continue
			}

			values[place] = dividend / divisor

			place++
		}

		places = place

		str = append(str, bm.chars[remainder])
	}

	// reverse the order of str
	l := len(str)
	for i := 0; i < l / 2; i++ {
		str[i], str[l-1-i] = str[l-1-i], str[i]
	}

	return string(str)
}

// NewBaseMap generates a base map
func NewBaseMap(chars string) (*BaseMap, error) {
	// use maps to make things faster
	vm := make(map[rune]int)
	cm := make(map[int]rune)

	var count int
	for i, c := range chars {
		_, ok := vm[c]
		if ok {
			return nil, ErrDuplicateCharacter
		}

		vm[c] = i
		cm[i] = c

		count++
	}

	return &BaseMap{radix: count, values: vm, chars: cm}, nil
}

// Convert converts the input from one radix to another
func Convert(input, fromBase, toBase string) (string, error) {
	bm1, err := NewBaseMap(fromBase)
	if err != nil {
		return "", err
	}

	bm2, err := NewBaseMap(toBase)
	if err != nil {
		return "", err
	}

	num, err := bm1.Parse(input)
	if err != nil {
		return "", err
	}

	return num.Format(bm2), nil
}
