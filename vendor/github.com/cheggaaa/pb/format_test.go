package pb

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func Test_DefaultsToInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := Format(value).String()

	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := Format(value).To(U_NO).String()

	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsBytes(t *testing.T) {
	inputs := []struct {
		v int64
		e string
	}{
		{v: 1000, e: "1000 B"},
		{v: 1024, e: "1.00 KiB"},
		{v: 3*MiB + 140*KiB, e: "3.14 MiB"},
		{v: 2 * GiB, e: "2.00 GiB"},
		{v: 2048 * GiB, e: "2.00 TiB"},
	}

	for _, input := range inputs {
		actual := Format(input.v).To(U_BYTES).String()
		if actual != input.e {
			t.Error(fmt.Sprintf("Expected {%s} was {%s}", input.e, actual))
		}
	}
}

func Test_CanFormatAsBytesDec(t *testing.T) {
	inputs := []struct {
		v int64
		e string
	}{
		{v: 999, e: "999 B"},
		{v: 1024, e: "1.02 KB"},
		{v: 3*MB + 140*KB, e: "3.14 MB"},
		{v: 2 * GB, e: "2.00 GB"},
		{v: 2048 * GB, e: "2.05 TB"},
	}

	for _, input := range inputs {
		actual := Format(input.v).To(U_BYTES_DEC).String()
		if actual != input.e {
			t.Error(fmt.Sprintf("Expected {%s} was {%s}", input.e, actual))
		}
	}
}

func Test_CanFormatDuration(t *testing.T) {
	value := 10 * time.Minute
	expected := "10m0s"
	actual := Format(int64(value)).To(U_DURATION).String()
	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_DefaultUnitsWidth(t *testing.T) {
	value := 10
	expected := "     10"
	actual := Format(int64(value)).Width(7).String()
	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}
