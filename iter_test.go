package iter

import (
	"fmt"
	"math/rand"
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

func BenchmarkEach(b *testing.B) {
	s := []string{"abc", "abd", "bcd"}
	it := New(FromStrings(s))

	each := func() {
		it.Each(func(v interface{}) {
			/* no i/o dep */
			_ = v.(string)
		})
	}
	forloop := func() {
		for i := 0; i < len(s); i++ {
			_ = s[i]
		}
	}

	tests := []struct {
		desc string
		run  func()
	}{
		{"each", each},
		{"loop", forloop},
	}

	for _, tc := range tests {
		b.Run(tc.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tc.run()
			}
		})
	}

	// Now setup a rather large random bits []string.
	n := 1 << 16
	const bitSize = 128
	s = s[:0]
	for i := 0; i < n; i++ {
		p := make([]byte, 0, bitSize)
		rand.Read(p)
		s = append(s, string(p))
	}

	it = New(FromStrings(s))

	for _, tc := range tests {
		b.Run(tc.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tc.run()
			}
		})
	}
}

func TestAdvance(t *testing.T) {
	it := New(FromStrings([]string{"a", "b", "c"}))
	n, more := it.Advance(2)
	if n != 1 || !more {
		t.Errorf("Advance(2) got index: %d and more: %t, but want: 1 and true.", n, more)
	}
	n, more = it.Advance(1)
	if n != 2 || !more {
		t.Errorf("Advance(1) after Advance(2) got index: %d and more: %t, but want: 2 and true.", n, more)
	}
	n, more = it.Advance(1)
	if n != 2 || more {
		t.Errorf("Advance(1) over the Iterator size got index: %d and more: %t, but want: 2 and false.", n, more)
	}

	// empty iter.
	it = New(FromStrings([]string{}))
	n, more = it.Advance(1)
	if n != 0 || more {
		t.Errorf("Advance(1) on an empty Iterator got index: %d and more: %t, but want: 0 and false.", n, more)
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		desc string
		it   *Iter
		size int
	}{
		{"empty", New(FromStrings([]string{})), 0},
		{"non-empty", New(FromStrings([]string{"a"})), 1},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.it.Count()
			if got != tc.size {
				t.Errorf("%s got count: %d but want: %d", tc.desc, got, tc.size)
			}
		})
	}

	it := New(FromStrings([]string{"a"}))
	n := it.Count()
	m := it.Count()
	if n != m {
		t.Errorf("Count multiple times yield invalid value. First time: %d, second time: %d", n, m)
	}
}
