package timefmt

import (
	"errors"
	"fmt"
	"time"
)

// Parse time string using the format.
func Parse(source, format string) (t time.Time, err error) {
	return parse(source, format, time.UTC, time.Local)
}

// ParseInLocation parses time string with the default location.
// The location is also used to parse the time zone name (%Z).
func ParseInLocation(source, format string, loc *time.Location) (t time.Time, err error) {
	return parse(source, format, loc, loc)
}

func parse(source, format string, loc, base *time.Location) (t time.Time, err error) {
	year, month, day, hour, min, sec, nsec := 1900, 1, 1, 0, 0, 0, 0
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse %q with %q: %w", source, format, err)
		}
	}()
	var j, century, yday, colons int
	var pm bool
	var pending string
	for i, l := 0, len(source); i < len(format); i++ {
		if b := format[i]; b == '%' {
			i++
			if i == len(format) {
				err = errors.New("stray %")
				return
			}
			b = format[i]
		L:
			switch b {
			case 'Y':
				if year, j, err = parseNumber(source, j, 4, 'Y'); err != nil {
					return
				}
			case 'y':
				if year, j, err = parseNumber(source, j, 2, 'y'); err != nil {
					return
				}
				if year < 69 {
					year += 2000
				} else {
					year += 1900
				}
			case 'C':
				if century, j, err = parseNumber(source, j, 2, 'C'); err != nil {
					return
				}
			case 'g':
				if year, j, err = parseNumber(source, j, 2, b); err != nil {
					return
				}
				year += 2000
			case 'G':
				if year, j, err = parseNumber(source, j, 4, b); err != nil {
					return
				}
			case 'm':
				if month, j, err = parseNumber(source, j, 2, 'm'); err != nil {
					return
				}
			case 'B':
				if month, j, err = lookup(source, j, longMonthNames, 'B'); err != nil {
					return
				}
			case 'b', 'h':
				if month, j, err = lookup(source, j, shortMonthNames, b); err != nil {
					return
				}
			case 'A':
				if _, j, err = lookup(source, j, longWeekNames, 'A'); err != nil {
					return
				}
			case 'a':
				if _, j, err = lookup(source, j, shortWeekNames, 'a'); err != nil {
					return
				}
			case 'w':
				if j >= l || source[j] < '0' || '6' < source[j] {
					err = parseFormatError(b)
					return
				}
				j++
			case 'u':
				if j >= l || source[j] < '1' || '7' < source[j] {
					err = parseFormatError(b)
					return
				}
				j++
			case 'V', 'U', 'W':
				if _, j, err = parseNumber(source, j, 2, b); err != nil {
					return
				}
			case 'e':
				if j < l && source[j] == ' ' {
					j++
				}
				fallthrough
			case 'd':
				if day, j, err = parseNumber(source, j, 2, b); err != nil {
					return
				}
			case 'j':
				if yday, j, err = parseNumber(source, j, 3, 'j'); err != nil {
					return
				}
			case 'k':
				if j < l && source[j] == ' ' {
					j++
				}
				fallthrough
			case 'H':
				if hour, j, err = parseNumber(source, j, 2, b); err != nil {
					return
				}
			case 'l':
				if j < l && source[j] == ' ' {
					j++
				}
				fallthrough
			case 'I':
				if hour, j, err = parseNumber(source, j, 2, b); err != nil {
					return
				}
				if hour == 12 {
					hour = 0
				}
			case 'p', 'P':
				var ampm int
				if ampm, j, err = lookup(source, j, []string{"AM", "PM"}, 'p'); err != nil {
					return
				}
				pm = ampm == 2
			case 'M':
				if min, j, err = parseNumber(source, j, 2, 'M'); err != nil {
					return
				}
			case 'S':
				if sec, j, err = parseNumber(source, j, 2, 'S'); err != nil {
					return
				}
			case 's':
				var unix int
				if unix, j, err = parseNumber(source, j, 10, 's'); err != nil {
					return
				}
				t = time.Unix(int64(unix), 0).In(time.UTC)
				var mon time.Month
				year, mon, day = t.Date()
				hour, min, sec = t.Clock()
				month = int(mon)
			case 'f':
				var msec, k, d int
				if msec, k, err = parseNumber(source, j, 6, 'f'); err != nil {
					return
				}
				nsec = msec * 1000
				for j, d = k, k-j; d < 6; d++ {
					nsec *= 10
				}
			case 'Z':
				k := j
				for ; k < l; k++ {
					if c := source[k]; c < 'A' || 'Z' < c {
						break
					}
				}
				t, err = time.ParseInLocation("MST", source[j:k], base)
				if err != nil {
					err = fmt.Errorf(`cannot parse %q with "%%Z"`, source[j:k])
					return
				}
				loc = t.Location()
				j = k
			case 'z':
				if j >= l {
					err = parseZFormatError(colons)
					return
				}
				sign := 1
				switch source[j] {
				case '-':
					sign = -1
					fallthrough
				case '+':
					var hour, min, sec, k int
					if hour, k, _ = parseNumber(source, j+1, 2, 'z'); k != j+3 {
						err = parseZFormatError(colons)
						return
					}
					if j = k; j >= l || source[j] != ':' {
						switch colons {
						case 1:
							err = errors.New("expected ':' for %:z")
							return
						case 2:
							err = errors.New("expected ':' for %::z")
							return
						}
					} else if j++; colons == 0 {
						colons = 4
					}
					if min, k, _ = parseNumber(source, j, 2, 'z'); k != j+2 {
						if colons == 0 {
							k = j
						} else {
							err = parseZFormatError(colons & 3)
							return
						}
					}
					if j = k; colons > 1 {
						if j >= l || source[j] != ':' {
							if colons == 2 {
								err = errors.New("expected ':' for %::z")
								return
							}
						} else if sec, k, _ = parseNumber(source, j+1, 2, 'z'); k != j+3 {
							if colons == 2 {
								err = parseZFormatError(colons)
								return
							}
						} else {
							j = k
						}
					}
					loc, colons = time.FixedZone("", sign*((hour*60+min)*60+sec)), 0
				case 'Z':
					loc, colons, j = time.UTC, 0, j+1
				default:
					err = parseZFormatError(colons)
					return
				}
			case ':':
				if pending != "" {
					if j >= l || source[j] != b {
						err = expectedFormatError(b)
						return
					}
					j++
				} else {
					if i++; i == len(format) {
						err = errors.New(`expected 'z' after "%:"`)
						return
					} else if b = format[i]; b == 'z' {
						colons = 1
					} else if b != ':' {
						err = errors.New(`expected 'z' after "%:"`)
						return
					} else if i++; i == len(format) {
						err = errors.New(`expected 'z' after "%::"`)
						return
					} else if b = format[i]; b == 'z' {
						colons = 2
					} else {
						err = errors.New(`expected 'z' after "%::"`)
						return
					}
					goto L
				}
			case 't', 'n':
				k := j
			K:
				for ; k < l; k++ {
					switch source[k] {
					case ' ', '\t', '\n', '\v', '\f', '\r':
					default:
						break K
					}
				}
				if k == j {
					err = fmt.Errorf("expected a space for %%%c", b)
					return
				}
				j = k
			case '%':
				if j >= l || source[j] != b {
					err = expectedFormatError(b)
					return
				}
				j++
			default:
				if pending == "" {
					var ok bool
					if pending, ok = compositions[b]; ok {
						break
					}
					err = fmt.Errorf(`unexpected format: "%%%c"`, b)
					return
				}
				if j >= l || source[j] != b {
					err = expectedFormatError(b)
					return
				}
				j++
			}
			if pending != "" {
				b, pending = pending[0], pending[1:]
				goto L
			}
		} else if j >= len(source) || source[j] != b {
			err = expectedFormatError(b)
			return
		} else {
			j++
		}
	}
	if j < len(source) {
		err = fmt.Errorf("unconverted string: %q", source[j:])
		return
	}
	if pm {
		hour += 12
	}
	if century > 0 {
		year = century*100 + year%100
	}
	if yday > 0 {
		return time.Date(year, time.January, 1, hour, min, sec, nsec, loc).AddDate(0, 0, yday-1), nil
	}
	return time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc), nil
}

type parseFormatError byte

func (err parseFormatError) Error() string {
	return fmt.Sprintf("cannot parse %%%c", byte(err))
}

type expectedFormatError byte

func (err expectedFormatError) Error() string {
	return fmt.Sprintf("expected %q", byte(err))
}

type parseZFormatError int

func (err parseZFormatError) Error() string {
	switch int(err) {
	case 0:
		return "cannot parse %z"
	case 1:
		return "cannot parse %:z"
	default:
		return "cannot parse %::z"
	}
}

func parseNumber(source string, min, size int, format byte) (int, int, error) {
	var val int
	if l := len(source); min+size > l {
		size = l
	} else {
		size += min
	}
	i := min
	for ; i < size; i++ {
		if b := source[i]; '0' <= b && b <= '9' {
			val = val*10 + int(b&0x0F)
		} else {
			break
		}
	}
	if i == min {
		return 0, 0, parseFormatError(format)
	}
	return val, i, nil
}

func lookup(source string, min int, candidates []string, format byte) (int, int, error) {
L:
	for i, xs := range candidates {
		j := min
		for k := 0; k < len(xs); k, j = k+1, j+1 {
			if j >= len(source) {
				continue L
			}
			if x, y := xs[k], source[j]; x != y && x|('a'-'A') != y|('a'-'A') {
				continue L
			}
		}
		return i + 1, j, nil
	}
	return 0, 0, parseFormatError(format)
}
