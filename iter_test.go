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

	t.Run("multi", func(t *testing.T) {
		// Count is rewinds for an Interable that implements Rewinder.
		it := New(FromStrings([]string{"a"}))
		n := it.Count()
		m := it.Count()
		if n != m {
			t.Errorf("Count multiple times yield invalid value. First time: %d, second time: %d", n, m)
		}

		it = New(FromStrings([]string{"a", "b"}))
		newit := it.Filter(func(v interface{}) bool { return v.(string) == "a" })
		c := newit.Count()
		if c != 1 {
			t.Errorf("Count after mutation function yields wrong value, got %d want 1", c)
		}
	})
}

func TestFunctions(t *testing.T) {
	tests := []struct {
		desc    string
		it      *Iter
		run     func(it *Iter) *Iter
		check   func(src, dst *Iter) error
		wantErr bool
	}{
		{
			"Filter-normal",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) *Iter { return it.Filter(func(v interface{}) bool { return v.(string) == "a" }) },
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 1 || o[0] != "a" {
					return fmt.Errorf("Filter %#+v incorrect. got: %#+v, want data []string{\"a\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Filter-empty",
			New(FromStrings([]string{})),
			func(it *Iter) *Iter { return it.Filter(func(v interface{}) bool { return v.(string) == "a" }) },
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 0 {
					return fmt.Errorf("Filter %#+v incorrect. got: %#+v, want data []string{nil}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Filter-everything",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) *Iter { return it.Filter(func(v interface{}) bool { return len(v.(string)) == 2 }) },
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 0 {
					return fmt.Errorf("Filter %#+v incorrect. got: %#+v, want data []string{nil}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Filter-nothing",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) *Iter { return it.Filter(func(v interface{}) bool { return len(v.(string)) == 1 }) },
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 2 || o[0] != "a" || o[1] != "b" {
					return fmt.Errorf("Filter %#+v incorrect. got: %#+v, want data []string{\"a\", \"b\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Map-normal",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) *Iter {
				return it.Map(func(v interface{}) interface{} { return strings.ToUpper(v.(string)) })
			},
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 2 || o[0] != "A" || o[1] != "B" {
					return fmt.Errorf("Map %#+v incorrect. got: %#+v, want data []string{\"A\", \"B\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Map-conditional",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) *Iter {
				return it.Map(func(v interface{}) interface{} {
					if v.(string) == "a" {
						return strings.ToUpper(v.(string))
					}
					return v
				})
			},
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 2 || o[0] != "A" || o[1] != "b" {
					return fmt.Errorf("Map %#+v incorrect. got: %#+v, want data []string{\"A\", \"b\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Every-even",
			New(FromStrings([]string{"a", "b", "c", "d"})),
			func(it *Iter) *Iter {
				return it.Every(func(i int, v interface{}) interface{} {
					if i%2 == 0 {
						return strings.ToUpper(v.(string))
					}
					return v
				})
			},
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 4 || o[0] != "A" || o[1] != "b" || o[2] != "C" || o[3] != "d" {
					return fmt.Errorf("Every %#+v incorrect. got: %#+v, want data []string{\"A\", \"b\", \"C\", \"d\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
		{
			"Or-strconv-fail",
			New(FromStrings([]string{"a", "1", "c", "2"})),
			func(it *Iter) *Iter {
				return it.Or(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				}, "not a number")
			},
			func(src, dst *Iter) error {
				o := dst.Collect().([]string)
				if len(o) != 4 || o[0] != "not a number" || o[2] != "not a number" {
					return fmt.Errorf("Or %#+v incorrect. got: %#+v, want data []string{\"not a number\", \"1\", \"not a number\", \"2\"}",
						src.impl.item, dst.impl.item)
				}
				return nil
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			it := tc.run(tc.it)
			if err := tc.check(tc.it, it); (err != nil) != tc.wantErr {
				t.Error(err)
			}
		})
	}
}

func TestFirstLast(t *testing.T) {
	tests := []struct {
		desc      string
		it        *Iter
		run       func(*Iter) (int, interface{}, bool)
		wantIdx   int
		wantValue interface{}
		wantMore  bool
	}{
		{
			"First-normal",
			New(FromStrings([]string{"a", "1", "b", "2"})),
			func(it *Iter) (int, interface{}, bool) {
				return it.First(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			1,
			"1",
			true,
		},
		{
			"First-empty",
			New(FromStrings([]string{})),
			func(it *Iter) (int, interface{}, bool) {
				return it.First(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			-1,
			nil,
			false,
		},
		{
			"First-nomatch",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) (int, interface{}, bool) {
				return it.First(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			-1,
			nil,
			false,
		},
		{
			"Last-normal",
			New(FromStrings([]string{"a", "1", "b", "2"})),
			func(it *Iter) (int, interface{}, bool) {
				return it.Last(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			3,
			"2",
			true,
		},
		{
			"Last-empty",
			New(FromStrings([]string{})),
			func(it *Iter) (int, interface{}, bool) {
				return it.Last(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			-1,
			nil,
			false,
		},
		{
			"Last-nomatch",
			New(FromStrings([]string{"a", "b"})),
			func(it *Iter) (int, interface{}, bool) {
				return it.Last(func(v interface{}) bool {
					_, err := strconv.Atoi(v.(string))
					return err == nil
				})
			},
			-1,
			nil,
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			idx, v, more := tc.run(tc.it)
			if idx != tc.wantIdx {
				t.Errorf("%s got item index of %d but want: %d", tc.desc, idx, tc.wantIdx)
			}

			if (v == nil) && tc.wantValue != nil {
				t.Errorf("%s got item nil value: but want: %v", tc.desc, tc.wantValue)
			}

			if (v != nil) && tc.wantValue == nil {
				t.Errorf("%s got item value: %v but want nil value.", tc.desc, v)
			}

			if v != nil && v.(string) != tc.wantValue.(string) {
				t.Errorf("%s got item value: %v but want: %v", tc.desc, v, tc.wantValue)
			}

			if more != tc.wantMore {
				t.Errorf("%s got more:%t but want: %t", tc.desc, more, tc.wantMore)
			}
		})
	}
}
