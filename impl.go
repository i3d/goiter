package iter

import (
	"fmt"
	"sync/atomic"
)

type iter struct {
	item Iterable
}

func newIter(item Iterable) *iter {
	return &iter{item}
}

func (it *iter) Filter(f FilterFunc) *iter {
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

func (it *iter) Map(f MapFunc) *iter {
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

func (it *iter) Each(f EachFunc) {
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

func (it *iter) Every(f EveryFunc) *iter {
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

func (it *iter) Or(f FilterFunc, this interface{}) *iter {
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

func (it *iter) Into(target Iterable, as ConvertFunc) *iter {
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

func (it *iter) From(other Iterable, as ConvertFunc) *iter {
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

func (it *iter) Nth(n int) interface{} {
	defer func() {
		if ag, ok := it.item.(Rewinder); ok {
			ag.Rewind()
		}
	}()

	for {
		i, v, more := it.item.(Enumerator).Enumerate()
		if !more {
			break
		}
		if i == n {
			return v
		}
	}
	return nil
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
