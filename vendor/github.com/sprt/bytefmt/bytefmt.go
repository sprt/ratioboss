// Package bytefmt provides utility functions to format and parse byte quantities.
package bytefmt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type unit int64

const (
	byt unit = 1 << (10 * iota)

	kibibyte
	mebibyte
	gibibyte
	tebibyte
	pebibyte
	exbibyte

	kilobyte = 1e3
	megabyte = 1e6
	gigabyte = 1e9
	terabyte = 1e12
	petabyte = 1e15
	exabyte  = 1e18
)

var symbols = map[unit]string{
	byt: "B",

	kibibyte: "KiB",
	mebibyte: "MiB",
	gibibyte: "GiB",
	tebibyte: "TiB",
	pebibyte: "PiB",
	exbibyte: "EiB",

	kilobyte: "KB",
	megabyte: "MB",
	gigabyte: "GB",
	terabyte: "TB",
	petabyte: "PB",
	exabyte:  "EB",
}

var suffixes = map[string]unit{
	"k": kibibyte,
	"m": mebibyte,
	"g": gibibyte,
	"t": tebibyte,
	"p": pebibyte,
	"e": exbibyte,

	"kb": kilobyte,
	"mb": megabyte,
	"gb": gigabyte,
	"tb": terabyte,
	"pb": petabyte,
	"eb": exabyte,
}

// ByteSize represents a quantity in bytes.
type ByteSize int64

func (s ByteSize) Binary() fmt.Formatter {
	return &formatter{s, kibibyte}
}

func (s ByteSize) SI() fmt.Formatter {
	return &formatter{s, kilobyte}
}

func (s *ByteSize) Set(v string) error {
	x, err := ParseCLI(v)
	if err != nil {
		return err
	}
	*s = x
	return nil
}

func (s *ByteSize) String() string {
	return strconv.FormatInt(int64(*s), 10)
}

func (s ByteSize) format(un unit) (float64, string) {
	u, ss := byt, unit(s)
	for ss >= un {
		ss /= un
		u *= un
	}
	return float64(s) / float64(u), symbols[u]
}

type formatter struct {
	bs ByteSize
	un unit
}

func (f *formatter) Format(s fmt.State, verb rune) {
	n, un := f.bs.format(f.un)
	prec, ok := s.Precision()
	if !ok {
		prec = -1
	}
	var b []byte
	b = strconv.AppendFloat(b, n, 'f', prec, 64)
	b = append(b, ' ')
	b = append(b, un...)
	s.Write(b)
}

// ParseCLI parses s and returns the corresponding size in bytes.
// s is a number followed by a unit (optional).
// Units are K,M,G,T,P,E,Z,Y (powers of 1024) and KB,MB,... (powers of 1000).
// ParseCLI is not case sensitive.
func ParseCLI(s string) (ByteSize, error) {
	s = strings.ToLower(s)
	for suffix, size := range suffixes {
		if strings.HasSuffix(s, suffix) {
			x, err := parseFloat(strings.TrimSuffix(s, suffix))
			if err != nil {
				return 0, err
			}
			return ByteSize(x * float64(size)), nil
		}
	}
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(x, 0) || math.IsNaN(x) {
		return 0, fmt.Errorf("cannot parse %q", s)
	}
	return ByteSize(x), nil
}

func parseFloat(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(x, 0) || math.IsNaN(x) {
		return 0, fmt.Errorf("cannot parse %q", s)
	}
	return x, nil
}
