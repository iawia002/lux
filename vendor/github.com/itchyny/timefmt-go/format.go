package timefmt

import (
	"strconv"
	"time"
)

// Format time to string using the format.
func Format(t time.Time, format string) string {
	return string(AppendFormat(make([]byte, 0, 64), t, format))
}

// AppendFormat appends formatted time to the bytes.
// You can use this method to reduce allocations.
func AppendFormat(buf []byte, t time.Time, format string) []byte {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var width, colons int
	var padding byte
	var pending string
	var upper, swap bool
	for i := 0; i < len(format); i++ {
		if b := format[i]; b == '%' {
			if i++; i == len(format) {
				buf = append(buf, '%')
				break
			}
			b, width, padding, upper, swap = format[i], 0, '0', false, false
		L:
			switch b {
			case '-':
				if pending != "" {
					buf = append(buf, '-')
					break
				}
				if i++; i == len(format) {
					goto K
				}
				padding = ^paddingMask
				b = format[i]
				goto L
			case '_':
				if i++; i == len(format) {
					goto K
				}
				padding = ' ' | ^paddingMask
				b = format[i]
				goto L
			case '^':
				if i++; i == len(format) {
					goto K
				}
				upper = true
				b = format[i]
				goto L
			case '#':
				if i++; i == len(format) {
					goto K
				}
				swap = true
				b = format[i]
				goto L
			case '0':
				if i++; i == len(format) {
					goto K
				}
				padding = '0' | ^paddingMask
				b = format[i]
				goto L
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				width = int(b & 0x0F)
				const maxWidth = 1024
				for i++; i < len(format); i++ {
					b = format[i]
					if b <= '9' && '0' <= b {
						width = width*10 + int(b&0x0F)
						if width >= int((^uint(0)>>1)/10) {
							width = maxWidth
						}
					} else {
						break
					}
				}
				if width > maxWidth {
					width = maxWidth
				}
				if padding == ^paddingMask {
					padding = ' ' | ^paddingMask
				}
				if i == len(format) {
					goto K
				}
				goto L
			case 'Y':
				if width == 0 {
					width = 4
				}
				buf = appendInt(buf, year, width, padding)
			case 'y':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, year%100, width, padding)
			case 'C':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, year/100, width, padding)
			case 'g':
				if width < 2 {
					width = 2
				}
				year, _ := t.ISOWeek()
				buf = appendInt(buf, year%100, width, padding)
			case 'G':
				if width == 0 {
					width = 4
				}
				year, _ := t.ISOWeek()
				buf = appendInt(buf, year, width, padding)
			case 'm':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, int(month), width, padding)
			case 'B':
				buf = appendString(buf, longMonthNames[month-1], width, padding, upper, swap)
			case 'b', 'h':
				buf = appendString(buf, shortMonthNames[month-1], width, padding, upper, swap)
			case 'A':
				buf = appendString(buf, longWeekNames[t.Weekday()], width, padding, upper, swap)
			case 'a':
				buf = appendString(buf, shortWeekNames[t.Weekday()], width, padding, upper, swap)
			case 'w':
				for ; width > 1; width-- {
					buf = append(buf, padding&paddingMask)
				}
				buf = append(buf, '0'+byte(t.Weekday()))
			case 'u':
				w := int(t.Weekday())
				if w == 0 {
					w = 7
				}
				for ; width > 1; width-- {
					buf = append(buf, padding&paddingMask)
				}
				buf = append(buf, '0'+byte(w))
			case 'V':
				if width < 2 {
					width = 2
				}
				_, week := t.ISOWeek()
				buf = appendInt(buf, week, width, padding)
			case 'U':
				if width < 2 {
					width = 2
				}
				week := (t.YearDay() + 6 - int(t.Weekday())) / 7
				buf = appendInt(buf, week, width, padding)
			case 'W':
				if width < 2 {
					width = 2
				}
				week := t.YearDay()
				if int(t.Weekday()) > 0 {
					week -= int(t.Weekday()) - 7
				}
				week /= 7
				buf = appendInt(buf, week, width, padding)
			case 'e':
				if padding < ^paddingMask {
					padding = ' '
				}
				fallthrough
			case 'd':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, day, width, padding)
			case 'j':
				if width < 3 {
					width = 3
				}
				buf = appendInt(buf, t.YearDay(), width, padding)
			case 'k':
				if padding < ^paddingMask {
					padding = ' '
				}
				fallthrough
			case 'H':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, hour, width, padding)
			case 'l':
				if width < 2 {
					width = 2
				}
				if padding < ^paddingMask {
					padding = ' '
				}
				h := hour
				if h > 12 {
					h -= 12
				}
				buf = appendInt(buf, h, width, padding)
			case 'I':
				if width < 2 {
					width = 2
				}
				h := hour
				if h > 12 {
					h -= 12
				} else if h == 0 {
					h = 12
				}
				buf = appendInt(buf, h, width, padding)
			case 'p':
				if hour < 12 {
					buf = appendString(buf, "AM", width, padding, upper, swap)
				} else {
					buf = appendString(buf, "PM", width, padding, upper, swap)
				}
			case 'P':
				if hour < 12 {
					buf = appendString(buf, "am", width, padding, upper, swap)
				} else {
					buf = appendString(buf, "pm", width, padding, upper, swap)
				}
			case 'M':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, min, width, padding)
			case 'S':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, sec, width, padding)
			case 's':
				if padding < ^paddingMask {
					padding = ' '
				}
				buf = appendInt(buf, int(t.Unix()), width, padding)
			case 'f':
				if width == 0 {
					width = 6
				}
				buf = appendInt(buf, t.Nanosecond()/1000, width, padding)
			case 'Z', 'z':
				name, offset := t.Zone()
				if b == 'Z' && name != "" {
					buf = appendString(buf, name, width, padding, upper, swap)
					break
				}
				i := len(buf)
				if padding != ^paddingMask {
					for ; width > 1; width-- {
						buf = append(buf, padding&paddingMask)
					}
				}
				j := len(buf)
				if offset < 0 {
					buf = append(buf, '-')
					offset = -offset
				} else {
					buf = append(buf, '+')
				}
				k := len(buf)
				buf = appendInt(buf, offset/3600, 2, padding)
				if buf[k] == ' ' {
					buf[k-1], buf[k] = buf[k], buf[k-1]
				}
				if k = offset % 3600; colons <= 2 || k != 0 {
					if colons != 0 {
						buf = append(buf, ':')
					}
					buf = appendInt(buf, k/60, 2, '0')
					if k %= 60; colons == 2 || colons == 3 && k != 0 {
						buf = append(buf, ':')
						buf = appendInt(buf, k, 2, '0')
					}
				}
				colons = 0
				if i != j {
					l := len(buf)
					k = j + 1 - (l - j)
					if k < i {
						l = j + 1 + i - k
						k = i
					} else {
						l = j + 1
					}
					copy(buf[k:], buf[j:])
					buf = buf[:l]
					if padding&paddingMask == '0' {
						for ; k > i; k-- {
							buf[k-1], buf[k] = buf[k], buf[k-1]
						}
					}
				}
			case ':':
				if pending != "" {
					buf = append(buf, ':')
				} else {
					colons = 1
				M:
					for i++; i < len(format); i++ {
						switch format[i] {
						case ':':
							colons++
						case 'z':
							if colons > 3 {
								i++
								break M
							}
							b = 'z'
							goto L
						default:
							break M
						}
					}
					buf = appendLast(buf, format[:i], width, padding)
					i--
					colons = 0
				}
			case 't':
				buf = appendString(buf, "\t", width, padding, false, false)
			case 'n':
				buf = appendString(buf, "\n", width, padding, false, false)
			case '%':
				buf = appendString(buf, "%", width, padding, false, false)
			default:
				if pending == "" {
					var ok bool
					if pending, ok = compositions[b]; ok {
						swap = false
						break
					}
					buf = appendLast(buf, format[:i], width-1, padding)
				}
				buf = append(buf, b)
			}
			if pending != "" {
				b, pending, width, padding = pending[0], pending[1:], 0, '0'
				goto L
			}
		} else {
			buf = append(buf, b)
		}
	}
	return buf
K:
	return appendLast(buf, format, width, padding)
}

func appendInt(buf []byte, num, width int, padding byte) []byte {
	if padding != ^paddingMask {
		padding &= paddingMask
		switch width {
		case 2:
			if num < 10 {
				buf = append(buf, padding)
				goto L1
			} else if num < 100 {
				goto L2
			} else if num < 1000 {
				goto L3
			} else if num < 10000 {
				goto L4
			}
		case 4:
			if num < 1000 {
				buf = append(buf, padding)
				if num < 100 {
					buf = append(buf, padding)
					if num < 10 {
						buf = append(buf, padding)
						goto L1
					}
					goto L2
				}
				goto L3
			} else if num < 10000 {
				goto L4
			}
		default:
			i := len(buf)
			for ; width > 1; width-- {
				buf = append(buf, padding)
			}
			j := len(buf)
			buf = strconv.AppendInt(buf, int64(num), 10)
			l := len(buf)
			if j+1 == l || i == j {
				return buf
			}
			k := j + 1 - (l - j)
			if k < i {
				l = j + 1 + i - k
				k = i
			} else {
				l = j + 1
			}
			copy(buf[k:], buf[j:])
			return buf[:l]
		}
	}
	if num < 100 {
		if num < 10 {
			goto L1
		}
		goto L2
	} else if num < 10000 {
		if num < 1000 {
			goto L3
		}
		goto L4
	}
	return strconv.AppendInt(buf, int64(num), 10)
L4:
	buf = append(buf, byte(num/1000)|'0')
	num %= 1000
L3:
	buf = append(buf, byte(num/100)|'0')
	num %= 100
L2:
	buf = append(buf, byte(num/10)|'0')
	num %= 10
L1:
	return append(buf, byte(num)|'0')
}

func appendString(buf []byte, str string, width int, padding byte, upper, swap bool) []byte {
	if width > len(str) && padding != ^paddingMask {
		if padding < ^paddingMask {
			padding = ' '
		} else {
			padding &= paddingMask
		}
		for width -= len(str); width > 0; width-- {
			buf = append(buf, padding)
		}
	}
	switch {
	case swap:
		if str[len(str)-1] < 'a' {
			for _, b := range []byte(str) {
				buf = append(buf, b|0x20)
			}
			break
		}
		fallthrough
	case upper:
		for _, b := range []byte(str) {
			buf = append(buf, b&0x5F)
		}
	default:
		buf = append(buf, str...)
	}
	return buf
}

func appendLast(buf []byte, format string, width int, padding byte) []byte {
	for i := len(format) - 1; i >= 0; i-- {
		if format[i] == '%' {
			buf = appendString(buf, format[i:], width, padding, false, false)
			break
		}
	}
	return buf
}

const paddingMask byte = 0x7F

var longMonthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longWeekNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortWeekNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var compositions = map[byte]string{
	'c': "a b e H:M:S Y",
	'+': "a b e H:M:S Z Y",
	'F': "Y-m-d",
	'D': "m/d/y",
	'x': "m/d/y",
	'v': "e-b-Y",
	'T': "H:M:S",
	'X': "H:M:S",
	'r': "I:M:S p",
	'R': "H:M",
}
