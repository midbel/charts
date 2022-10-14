package dsl

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

func makeTimeFormat(format string) (func(time.Time) string, error) {
	format, err := parseFormat(format)
	if err != nil {
		return nil, err
	}
	return func(t time.Time) string {
		return t.Format(format)
	}, nil
}

func makeParseTime(format string) (func(string) (time.Time, error), error) {
	format, err := parseFormat(format)
	if err != nil {
		return nil, err
	}
	return func(str string) (time.Time, error) {
		return time.Parse(format, str)
	}, nil
}

const percent = '%'

var specifiers = map[rune]string{
	'D': "01/02/06", // month/day/year
	'Y': "2006",     // year four digits
	'y': "06",       // year two digits
	'm': "01",       // month two digits
	'B': "January",  // full month name
	'b': "Jan",      // abreviate month name
	'h': "Jan",      // abreviate month name
	'd': "02",       // day of month
	'e': "_2",       // day of month space padded
	'j': "002",      // day of year
	'A': "Monday",   // full week day name
	'a': "Mon",      // abreviate week day name
	'H': "15",       // hours 00-23
	'I': "03",       // hours 00-12
	'k': "_3",       // hours 00-23 space padded
	'M': "04",       // minute two digits
	'S': "05",       // second two digits
	'p': "PM",       //
	'T': "15:04:05",
	'F': "2006-01-02",
	'z': "-07:00",
	'Z': "Z",
	'c': "Mon Jan 2 15:04:05 2006",
	'r': "03:04:05 PM",
	'R': "15:04",
	'%': "%",
	'n': "\n",
	't': "\t",
}

func parseFormat(str string) (string, error) {
	var (
		r = strings.NewReader(str)
		w strings.Builder
	)
	for r.Len() > 0 {
		x, _, _ := r.ReadRune()
		if x == utf8.RuneError {
			return "", fmt.Errorf("invalid character found in format string")
		}
		if x != percent {
			w.WriteRune(x)
			continue
		}
		x, _, _ = r.ReadRune()
		str, ok := specifiers[x]
		if !ok {
			return "", fmt.Errorf("invalid specifier found %c", x)
		}
		w.WriteString(str)
	}
	return w.String(), nil
}
