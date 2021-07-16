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
		if newelm, err := as(elm); err == nil {
			target.Add(newelm)
		}
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
		if thiselm, err := as(elm); err == nil {
			newitem.Add(thiselm)
		}
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
	var idx int = -1
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

func (it *iter) chain(other Iterable) *iter {
	newit, err := it.item.New()
	if err != nil {
		panic(err)
	}

	for {
		v, more := it.item.Next()
		if !more {
			break
		}
		newit.Add(v)
	}

	for {
		v, more := other.Next()
		if !more {
			break
		}
		newit.Add(v)
	}

	return newIter(newit)
}

func (it *iter) zip(other Iterable) *iter {
	np, _ := newPairs()

	for {
		v1, more1 := it.item.Next()
		v2, more2 := other.Next()
		if !more1 || !more2 {
			break
		}
		p := &Pair{v1, v2}
		np.Add(p)
	}
	return newIter(np)
}

type pairs struct {
	idx  int
	data []*Pair
	size int
}

func newPairs() (Iterable, error) {
	return &pairs{idx: -1}, nil
}

func (*pairs) New() (Iterable, error) {
	return newPairs()
}

func (ps *pairs) Next() (interface{}, bool) {
	ps.idx++
	if ps.idx < ps.size {
		return ps.data[ps.idx], true
	}
	return nil, false
}

func (ps *pairs) Rewind() {
	ps.idx = -1
}

func (ps *pairs) Reset() {
	ps.Rewind()
	ps.data = nil
	ps.size = 0
}

// Add inserts an string as an interface into the pairs struct.
func (ps *pairs) Add(obj interface{}) {
	input := obj.(*Pair)
	ps.data = append(ps.data, input)
	ps.size++
}

// Enumerate returns a pair of {index, string as interface}
// as well as a bool to indicate whether there ps more to go.
func (ps *pairs) Enumerate() (int, interface{}, bool) {
	ps.idx++
	if ps.idx < ps.size {
		return ps.idx, ps.data[ps.idx], true
	}
	return -1, nil, false
}

// To returns the underlying []*Pair back.
func (ps *pairs) To() interface{} {
	return ps.data
}

// String implements the Stringer interface for pairs.
func (ps *pairs) String() string {
	return fmt.Sprintf("%+v", ps.data)
}

// String provides a stringify impl for Pair.
func (p *Pair) String() string {
	return fmt.Sprintf("{%+v, %+v}", p.X, p.Y)
}

// === internal for testing ===

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
