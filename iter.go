// Package iter implements a common Iterable API.
// See tests for common usage.
// This package provides some similar functionality comparing to
// Rust's Iterator trait.
//
// This package might not yet be very interesting until Go's generics
// is out, at which moment, it should be easy to implement Iterable
// for type []T and reimplement the utility functions.
package iter

import (
	"fmt"
	"sync/atomic"
)

// Iterable is capable of traversing
// elements from some kind of collection.
type Iterable interface {
	// New initializes a new Iterable instance.
	New() (Iterable, error)
	// Add pushes an item into the existing Iterable.
	Add(interface{})
	// Next emits an item from the existing Iterable.
	// The second return as a bool indicates whether
	// there is any more items expected.
	// Calling Next() when the bool value is false yields
	// undefined behavior.
	Next() (interface{}, bool)
}

// Enumerator is capable of traversing
// elements and their indexes from some kind of collection.
type Enumerator interface {
	Enumerate() (int, interface{}, bool)
}

// Rewinder can rewind the traversal back to a previous
// state so that the Iterable can traverse immeidately
// again. An Iterable doesn't implement Rewinder can't not
// be used after all items are traversed.
type Rewinder interface {
	Rewind()
}

// Nther can retrieve the n'th item from an Iterable.
type Nther interface {
	Nth(int) interface{}
}

// Resetter resets an Iterable to its initial state.
type Resetter interface {
	Reset()
}

// Intoer converts an Iterable with type T to another
// Iterable with type U.
// If the target Iterable is a Resetter, an Into implementation
// may call Reset before Add any items, otherwise,
// Into assumes the target is ready.
type Intoer interface {
	// Into assumes a newly initialized target Iterable
	// as its first argument.
	Into(Iterable, ConvertFunc) *Iter
}

// Fromer converts an Iterable with type U to itself.
// If this Iterable is a Reseter, a From implementation
// may call Reset on itself before Add any items, otherwise,
// it may return a new Iterable by calling the New interface.
type Fromer interface {
	// Fromer assumes the Iterable from its first argument
	// is the one to convert from.
	From(Iterable, ConvertFunc) *Iter
}

// FilterFunc runs a function with an given item and return a bool
// indicates some sort of predicates.
type FilterFunc func(interface{}) bool

// MapFunc applies some logic to an given item and returns a new
// (or same) item with the same underlying type.
type MapFunc func(interface{}) interface{}

// ConvertFunc is like MapFunc but converts T to U or back and forth.
type ConvertFunc func(interface{}) interface{}

// EachFunc runs a function on a given item without changin the state
// of that item.
type EachFunc func(interface{})

// EveryFunc runs a function on a give {index, item} pair and return
// a new (or same) item for that index.
type EveryFunc func(int, interface{}) interface{}

// Iter implements common utility functions for an Iterable.
type Iter struct {
	impl *iter
}

// New creates a new Iter.
func New(some Iterable) *Iter {
	return newFromImpl(newIter(some))
}

func newFromImpl(impl *iter) *Iter {
	return &Iter{impl}
}

// Filter applies a given predicate against every element of the Iterable
// and return a new Iterator that contains only items which the predicate
// returned true.
func (it *Iter) Filter(f FilterFunc) *Iter {
	return newFromImpl(it.impl.Filter(f))
}

// Map applies a given function (often mutation) against every item of
// the Iterable and return a new Iterator contains those (often mutated)
// items.
func (it *Iter) Map(f MapFunc) *Iter {
	return newFromImpl(it.impl.Map(f))
}

// Every applies a given function (often mutation) with a pair of (index, item)
// for every item of the Iterable and return a new Iterable contains those
// (often mutated) items.
// Every requires the underlying Iterable also is an Enumerator.
func (it *Iter) Every(f EveryFunc) *Iter {
	return newFromImpl(it.impl.Every(f))
}

// Or applies a given predicate for every item of an Iterable. If the predicate
// returns true, the item is not chagned, otherwise, the given item will be used
// to replace the existing item, serving like a default value.
func (it *Iter) Or(f FilterFunc, this interface{}) *Iter {
	return newFromImpl(it.impl.Or(f, this))
}

// Nth returns the n'th item from the Iterable.
// If N isn't in the valid iteration scope, nil will be returned.
// If the Iterable is also a Rewinder, then after retrieving
// the Nth item, the Iterable will be rewinded and assumed to be
// reusable immeidately.
func (it *Iter) Nth(n int) interface{} {
	return it.impl.Nth(n)
}

// Each runs a function against each item for an Iterable
// without changing the item state.
// If the Iterable is also a Rewinder, then after iterating
// all items, the Iterable will be rewinded and assumed to be
// reusable immeidately.
func (it *Iter) Each(f EachFunc) {
	it.impl.Each(f)
}

// Into converts self Iterable with underlying type T to another Iterable
// with underlying type U.
// If other is a Resetter, then Reset will be called, otherwise
// assume other is clean.
func (it *Iter) Into(target Iterable, as ConvertFunc) *Iter {
	return newFromImpl(it.impl.Into(target, as))
}

// From converts other Iterable with type U to self with type T.
// If self is a Resetter, then Reset will be called, otherwise,
// assume clean.
func (it *Iter) From(other Iterable, as ConvertFunc) *Iter {
	return newFromImpl(it.impl.From(other, as))
}

// Iterator for []string.
// This is the only Iterable implementation provided by the API
// since Go hasn't yet had Generics. It would be tedious if not
// impossible to implement all []T. So if there is a need for
// some T, client will have to implement on thir own.

// IterStrings implements Iterable API for []string.
type IterStrings struct {
	idx  int32
	data []string
}

// NewIterStrings creates a new empty IterStrings struct.
func NewIterStrings() *IterStrings {
	return &IterStrings{idx: -1}
}

// FromStrings creates a new IterStrings from a []string.
func FromStrings(s []string) *IterStrings {
	return &IterStrings{idx: -1, data: s}
}

// New constructs a new empty IterStrings from itself.
func (is *IterStrings) New() (Iterable, error) {
	return NewIterStrings(), nil
}

// Next returns the next string in iterator as an interface{}.
// bool indicate whether there is any more to go. If false,
// then this Iterator is exhausted.
func (is *IterStrings) Next() (interface{}, bool) {
	need := atomic.AddInt32(&is.idx, 1)
	var more bool = true

	if need > int32(len(is.data)-1) {
		more = false
	}

	if !more {
		return nil, more
	}
	return is.data[need], more
}

// Rewind for IterStrings will set the Iterator to its initial
// traversal state and ready for start from beginning again.
func (is *IterStrings) Rewind() {
	atomic.StoreInt32(&is.idx, -1)
}

// Reset sets this IterStrings to it's initial state.
// Whatever data hosted would be lost after this call.
func (is *IterStrings) Reset() {
	is.Rewind()
	is.data = nil
}

// Add inserts an string as an interface into the IterStrings struct.
func (is *IterStrings) Add(obj interface{}) {
	input := obj.(string)
	is.data = append(is.data, input)
}

// Enumerate returns a pair of {index, string as interface}
// as well as a bool to indicate whether there is more to go.
func (is *IterStrings) Enumerate() (int, interface{}, bool) {
	need := atomic.AddInt32(&is.idx, 1)
	var more bool = true

	if need > int32(len(is.data)-1) {
		more = false
	}
	if !more {
		return -1, nil, more
	}
	return int(need), is.data[need], more
}

// String implements the Stringer interface for IterStrings.
func (is *IterStrings) String() string {
	return fmt.Sprintf("%+v", is.data)
}
