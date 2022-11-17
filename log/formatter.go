package log

import (
	"bytes"
	"fmt"
	"time"
)

type Formatter interface {
	Format(*Entry) ([]byte, error)
}

func timeFormatItoA(number int, wid int) []byte {
	var temp [10]byte
	moveIndex := len(temp) - 1
	for number >= 10 || wid > 1 {
		wid--
		temp[moveIndex] = byte('0' + (number % 10))
		moveIndex--
		number /= 10
	}
	temp[moveIndex] = byte('0' + number)
	return temp[moveIndex:]
}

type TextFormatter struct {
	DisableTimestamp bool
}

func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	b.WriteString(entry.Level.String())
	if !f.DisableTimestamp {
		f.TimeFormat(b, entry.Time)
	}
	var callerVal string
	if entry.HasCaller() {
		callerVal = fmt.Sprintf("[%s:%d:%s]", entry.Caller.File, entry.Caller.Line, entry.Caller.Function)
		b.WriteString(callerVal)
	}

	if entry.Message != "" {
		b.WriteString(" ")
		b.WriteString(entry.Message)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) TimeFormat(b *bytes.Buffer, t time.Time) {
	b.WriteByte('[')
	year, month, day := t.Date()
	yearStr := timeFormatItoA(year, 4)
	b.Write(yearStr)
	b.WriteByte('/')
	monthStr := timeFormatItoA(int(month), 2)
	b.Write(monthStr)
	b.WriteByte('/')
	dayStr := timeFormatItoA(day, 2)
	b.Write(dayStr)
	b.WriteByte(' ')
	hour, min, sec := t.Clock()
	hourStr := timeFormatItoA(hour, 2)
	b.Write(hourStr)
	b.WriteByte(':')
	minStr := timeFormatItoA(min, 2)
	b.Write(minStr)
	b.WriteByte(':')
	secStr := timeFormatItoA(sec, 2)
	b.Write(secStr)
	b.WriteByte('.')
	msStr := timeFormatItoA(t.Nanosecond()/1e6, 3)
	b.Write(msStr)
	b.WriteByte(']')
}

type NullFormatter struct {
}

func (f *NullFormatter) Format(entry *Entry) ([]byte, error) {
	return []byte{}, nil
}
