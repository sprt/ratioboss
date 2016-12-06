// Package bytefmt provides utility functions to format and parse byte
// quantities.
package byt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// A Size is a quantity in bytes. It implements the flag.Value interface.
type Size int64

// FlagUsage is a partial usage message that applications using a Size as a flag
// may wish to include in their -help output.
const FlagUsage = `
A byte size is a number and optional unit. Units are K,M,G,T,P,E,Z,Y (powers of
1024) or KB,MB,... (powers of 1000). It is not case sensitive.
`

const (
	Byt Size = 1 << (10 * iota)

	Kibibyte
	Mebibyte
	Gibibyte
	Tebibyte
	Pebibyte
	Exbibyte

	K = Kibibyte
	M = Mebibyte
	G = Gibibyte
	T = Tebibyte
	P = Pebibyte
	E = Exbibyte

	Kilobyte = 1e3
	Megabyte = 1e6
	Gigabyte = 1e9
	Terabyte = 1e12
	Petabyte = 1e15
	Exabyte  = 1e18

	KB = Kilobyte
	MB = Megabyte
	GB = Gigabyte
	TB = Terabyte
	PB = Petabyte
	EB = Exabyte
)

var symbols = map[Size]string{
	Byt: "B",

	Kibibyte: "KiB",
	Mebibyte: "MiB",
	Gibibyte: "GiB",
	Tebibyte: "TiB",
	Pebibyte: "PiB",
	Exbibyte: "EiB",

	Kilobyte: "KB",
	Megabyte: "MB",
	Gigabyte: "GB",
	Terabyte: "TB",
	Petabyte: "PB",
	Exabyte:  "EB",
}

var suffixes = map[string]Size{
	"k": Kibibyte,
	"m": Mebibyte,
	"g": Gibibyte,
	"t": Tebibyte,
	"p": Pebibyte,
	"e": Exbibyte,

	"kb": Kilobyte,
	"mb": Megabyte,
	"gb": Gigabyte,
	"tb": Terabyte,
	"pb": Petabyte,
	"eb": Exabyte,
}

func (s Size) Binary() fmt.Formatter {
	return &formatter{s, Kibibyte}
}

func (s Size) Decimal() fmt.Formatter {
	return &formatter{s, Kilobyte}
}

func (s *Size) Set(v string) error {
	x, err := parseCLI(v)
	if err != nil {
		return err
	}
	*s = x
	return nil
}

func (s *Size) String() string {
	return strconv.FormatInt(int64(*s), 10)
}

func (s Size) format(un Size) (float64, string) {
	u, ss := Byt, s
	for ss >= un {
		ss /= un
		u *= un
	}
	return float64(s) / float64(u), symbols[u]
}

type formatter struct {
	bs Size
	un Size
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

// parseCLI parses a flag value and returns the specified size;
// see FlagUsage for details.
func parseCLI(s string) (Size, error) {
	s = strings.ToLower(s)
	for suffix, size := range suffixes {
		if strings.HasSuffix(s, suffix) {
			x, err := parseFloat(strings.TrimSuffix(s, suffix))
			if err != nil {
				return 0, err
			}
			return Size(x * float64(size)), nil
		}
	}
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(x, 0) || math.IsNaN(x) {
		return 0, fmt.Errorf("cannot parse %q", s)
	}
	return Size(x), nil
}

func parseFloat(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(x, 0) || math.IsNaN(x) {
		return 0, fmt.Errorf("cannot parse %q", s)
	}
	return x, nil
}
