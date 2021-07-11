package iter

import (
	"fmt"
	"sync/atomic"
)

type iter struct {
	item Iterable
	size int
}

func newIter(item Iterable) *iter {
	return &iter{item: item}
}

func (it *iter) filter(f FilterFunc) *iter {
	newitem, err := it.item.New()
	if err != nil {
		panic(err)
	}

	for {
		elm, more := it.item.Next()
		if !more {
			break
		}
		if f(elm) {
			newitem.Add(elm)
		}
	}
	return newIter(newitem)
}

func (it *iter) apply(f MapFunc) *iter {
	newitem, err := it.item.New()
	if err != nil {
		panic(err)
	}

	for {
		elm, more := it.item.Next()
		if !more {
			break
		}
		newitem.Add(f(elm))
	}
	return newIter(newitem)
}

func (it *iter) each(f EachFunc) {
	defer func() {
		if ag, ok := it.item.(Rewinder); ok {
			ag.Rewind()
		}
	}()

	for {
		elm, more := it.item.Next()
		if !more {
			return
		}
		f(elm)
	}
}

func (it *iter) every(f EveryFunc) *iter {
	newitem, err := it.item.New()
	if err != nil {
		panic(err)
	}

	for {
		i, v, more := it.item.(Enumerator).Enumerate()
		if !more {
			break
		}
		newitem.Add(f(i, v))
	}
	return newIter(newitem)
}

func (it *iter) or(f FilterFunc, this interface{}) *iter {
	newitem, err := it.item.New()
	if err != nil {
		panic(err)
	}

	for {
		elm, more := it.item.Next()
		if !more {
			break
		}
		if f(elm) {
			newitem.Add(elm)
		} else {
			newitem.Add(this)
		}
	}
	return newIter(newitem)
}

func (it *iter) into(target Iterable, as ConvertFunc) *iter {
	if resetter, ok := target.(Resetter); ok {
		resetter.Reset()
	}

	for {
		elm, more := it.item.Next()
		if !more {
			break
		}
		newelm := as(elm)
		target.Add(newelm)
	}

	return newIter(target)
}

func (it *iter) from(other Iterable, as ConvertFunc) *iter {
	var newitem Iterable
	var newit *iter
	var err error

	if r, ok := it.item.(Resetter); ok {
		r.Reset()
		newitem = it.item
		newit = it
	} else {
		newitem, err = it.item.New()
		if err != nil {
			panic(err)
		}
		newit = newIter(newitem)
	}

	for {
		elm, more := other.Next()
		if !more {
			break
		}
		thiselm := as(elm)
		newitem.Add(thiselm)
	}
	return newit
}

func (it *iter) advanceBy(n int) (int, bool) {
	var more bool

	for i := 0; i < n; i++ {
		_, more = it.item.Next()
		if !more {
			break
		}
		it.size++
	}

	idx := it.size - 1
	if idx <= 0 {
		idx = 0
	}
	return idx, more
}

func (it *iter) count() int {
	defer func() {
		if ag, ok := it.item.(Rewinder); ok {
			ag.Rewind()
			it.size = 0
		}
	}()

	var more = true
	for more {
		_, more = it.advanceBy(1)
	}
	return it.size
}

func (it *iter) first(f FilterFunc) (int, interface{}, bool) {
	var i int
	var v interface{}
	var more = true

	// NOTE: consider implementing faster search algorithm.
	for {
		i, v, more = it.item.(Enumerator).Enumerate()
		if !more {
			break
		}
		if f(v) {
			break
		}
	}
	return i, v, more
}

func (it *iter) last(f FilterFunc) (int, interface{}, bool) {
	var idx int
	var seen interface{}
	var found bool

	// NOTE: consider implementing faster search algorithm.
	for {
		i, v, more := it.item.(Enumerator).Enumerate()
		if !more {
			break
		}
		if f(v) {
			found = true
			seen = v
			idx = i
		}
	}
	return idx, seen, found
}

// An internal Iterable impl for []int,
// used for Into/From conversion tests.
type iterInts struct {
	data []int
	idx  int64
}

func (is *iterInts) New() (Iterable, error) {
	return &iterInts{idx: -1}, nil
}

func (is *iterInts) Next() (interface{}, bool) {
	need := atomic.AddInt64(&is.idx, 1)
	var more bool = true

	if need > int64(len(is.data)-1) {
		more = false
	}

	if !more {
		return nil, more
	}
	return is.data[need], more
}

func (is *iterInts) Rewind() {
	atomic.StoreInt64(&is.idx, -1)
}

func (is *iterInts) Reset() {
	is.Rewind()
	is.data = nil
}

func (is *iterInts) Add(obj interface{}) {
	input := obj.(int)
	is.data = append(is.data, input)
}

func (is *iterInts) Enumerate() (int, interface{}, bool) {
	need := atomic.AddInt64(&is.idx, 1)
	var more bool = true

	if need > int64(len(is.data)-1) {
		more = false
	}
	if !more {
		return -1, nil, more
	}
	return int(need), is.data[need], more
}

func (is *iterInts) String() string {
	return fmt.Sprintf("%#+v", is.data)
}
