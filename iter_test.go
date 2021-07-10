package iter

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestIterString(t *testing.T) {
	s := []string{"abc", "bbc", "abccd", "abcdd"}
	it := New(FromStrings(s))

	newit := it.
		Filter(func(item interface{}) bool {
			elm := item.(string)
			return strings.HasPrefix(elm, "ab")
		}).
		Or(func(item interface{}) bool {
			return item.(string) != "abcdd"
		}, "abcde").
		Map(func(item interface{}) interface{} {
			return fmt.Sprintf("%s starts from 'ab'", item.(string))
		}).
		Every(func(i int, v interface{}) interface{} {
			return fmt.Sprintf("%d: %s", i, v.(string))
		})

	newit.Each(func(it interface{}) {
		fmt.Printf("%s\n", it)
	})

	// layer is now:
	// zero := "0: abc starts from 'ab'"
	// one := "1: abccd starts from 'ab'"
	two := "2: abcde starts from 'ab'"

	if newit.Nth(2) != two {
		t.Errorf("Nth element is wrong, got: %s, want:%s", newit.Nth(2), two)
	}
}

func TestFrom(t *testing.T) {
	d := []int{1, 2, 3}
	ints := &iterInts{d, -1}

	s := New(NewIterStrings())
	s.From(ints, func(v interface{}) interface{} {
		return fmt.Sprintf("%d", v)
	}).Each(func(v interface{}) {
		fmt.Printf("%s\n", v)
	})
}

func TestInto(t *testing.T) {
	s := New(FromStrings([]string{"1", "2", "3"}))
	ints := &iterInts{nil, -1}
	s.Into(ints, func(v interface{}) interface{} {
		i, err := strconv.Atoi(v.(string))
		if err != nil {
			t.Fatal(err)
		}
		return i
	})

	if ints.data[0] != 1 || ints.data[1] != 2 || ints.data[2] != 3 {
		t.Errorf("Into conversion is invalid, got: %+v, want: []int{1,2,3}", ints.data)
	}
}
