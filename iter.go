// Package iter implements a common Iterable API.
// See tests for common usage.
// This package provides some similar functionality comparing to
// Rust's Iterator trait.
//
// This package might not yet be very interesting until Go's generics
// is out, at which point, it should be easy to implement Iterable
// for type []T and reimplement the utility functions.
package iter

import (
	"fmt"
)

// Iterable is capable of traversing
// elements from some kind of collection.
//
// An implementation of Iterable can be used
// directly or, typically, be consumed by an
// Iterator taking advantage of the Iterable
// protocol.
//
// In this API, most of the mutation APIs
// from the Iterator yields a new Iterator
// instead of mutating the existing one,
// so we require an Iterable also provides
// New and Add interfaces.
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
// elements and their indexes from some kind of
// collection.
//
// In addition to the Next() API, if an Iterable
// also implements Enumerator, it is then able to
// traverse element with a pair of {index, value}.
// A collection with some ordering semantics can
// consider also implementing the Enumerator
// interface, which will unleach the Iterator
// doing some more powerful things.
type Enumerator interface {
	Enumerate() (int, interface{}, bool)
}

// Rewinder can rewind the traversal back to a previous
// state so that the same Iterable can traverse
// immeidately again.
//
// An Iterable doesn't implement Rewinder can't not
// be used after all items are traversed. This is commonly
// called "consumed". Without a Rewinder, even read-only
// APIs "consume" the Iterable.
type Rewinder interface {
	Rewind()
}

// Resetter resets an Iterable to its initial state.
// This is optional. For example, in order to take
// advantage of the Iterator's Into/From APIs, an Iterable
// shall consider implementing this interface so that
// when converting this Iterator wth Iterable type T into
// another Iterator with Iterable type U, or vice verse,
// the target Iterable can be correctly initialized.
type Resetter interface {
	Reset()
}

// Intoer converts an Iterator with Iterable type T
// to another Iterator with Iterable type U.
// If the target Iterable is a Resetter, an Intoer
// implementation may call Reset before Add any items,
// otherwise, Intoer shall assume the target is ready.
type Intoer interface {
	// Into assumes a newly initialized target Iterable
	// as its first argument.
	Into(Iterable, ConvertFunc) *Iter
}

// Fromer converts an Iterator with Iterable type U to
// itself (type T).
// If the underlying Iterable is a Reseter, a Fromer
// implementation may call Reset on itself before Add
// any items, otherwise, it may return a new Iterable
// by calling the New() API.
type Fromer interface {
	// Fromer assumes the Iterable from its first argument
	// is srouce Iterable to convert from.
	From(Iterable, ConvertFunc) *Iter
}

// FilterFunc runs a function with an given item and return a bool
// indicates some sort of predicates.
type FilterFunc func(interface{}) bool

// MapFunc applies some logic to an given item and returns a new
// (or same) item with the same underlying type.
type MapFunc func(interface{}) interface{}

// ConvertFunc likes the MapFunc but converts type T to U or
// back and forth.
type ConvertFunc func(interface{}) interface{}

// EachFunc runs a function on a given item without changin the state
// of that item.
type EachFunc func(interface{})

// EveryFunc runs a function on a give {index, item} pair and return
// a new (or same) item for that index.
type EveryFunc func(int, interface{}) interface{}

// Iter is an Iterator implements common utility functions
// for an Iterable.
//
// The Iterator APIs offered here are heavily inspired by Rust's
// Iterator traits. The goal is to provide some familiarity and
// similarity to these two languages. After all, common concepts
// and powerful functions are useful regardless what languages
// they are used.
//
// It is however NOT the goal to provide a 1:1 mapping of the Rust
// API because Go is quite a different language than Rust. Go's
// Iterator API shall do the things in Go's way. The most important
// thing here is to capture the common Iterator concepts.
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
//
// Example:
//   it := New(FromStrings([]string{"abc", "abd", "bcd"}))
//   newit := it.Filter(func(v interface{}) bool {
//      return v.(string) == "abc"
//   })
//   produces a newit contains []string{"abc"}
func (it *Iter) Filter(f FilterFunc) *Iter {
	return newFromImpl(it.impl.filter(f))
}

// Map applies a given function (often mutation) against every item of
// the Iterable and return a new Iterator contains those (often mutated)
// items.
//
// Example:
//   it := New(FromStrings([]string{"a", "b"}))
//   newit := it.Map(func(v interface{}) interface{} {
//     return fmt.Sprintf("%s seen")
//   })
//   produces a newit contains []string{"a seen", "b seen"}
func (it *Iter) Map(f MapFunc) *Iter {
	return newFromImpl(it.impl.apply(f))
}

// Every applies a given function (often mutation) with a pair of
// (index, item) for every item of the Iterable and return a new
// Iterator contains those (often mutated) items.
// Every requires the underlying Iterable also is an Enumerator.
//
// Example:
//   it := New(FromStrings([]string{"a", "b"}))
//   newit := it.Every(func(i int, v interface{}) interface{} {
//          if i % 2 == 0 {
//              return fmt.Sprintf("Even: %s", v)
//          }
//          return fmt.Sprintf("Odd: %s", v)
//   })
//   produces a newit contains []string{"Odd: a", "Even: b"}
func (it *Iter) Every(f EveryFunc) *Iter {
	return newFromImpl(it.impl.every(f))
}

// Or applies a given predicate for every item of an Iterable.
// If the predicate returns true, the item is not chagned,
// otherwise, the given item will be used to replace the existing
// item, serving like a default value.
//
// Example:
//   it := New(FromStrings([]string{"a", "b"}))
//   newit := it.Or(func(v interface{}) bool {
//     return v.(string) == "b"
//   }, "invalid")
//   produces a newit contains []string{"a", "invalid"}
func (it *Iter) Or(f FilterFunc, this interface{}) *Iter {
	return newFromImpl(it.impl.or(f, this))
}

// Advance moves the Iterable's item position forward by N times.
// If the underlying Iterable is index-based, this means the returned
// int points to index N-1 when N is a valid move.
// The returned bool indicates whether the Advance has exhausted
// the Iterable size (can it go further). If false, int guarantees
// point to the last index, in other words, calling Next() on the Iterable
// would be invalid. Obviously, when bool == false, int indicates the
// size of the Iterable.
//
// Example:
//   it := New(FromStrings([]string{"a,", "b"}))
//   it.Advance(1) => 0, true
//   it.Advance(1) => 1, true
//   it.Advance(1) => 1, false
//   it.Advance(5) => 1, false
func (it *Iter) Advance(n int) (int, bool) {
	return it.impl.advanceBy(n)
}

// Count returns the size of the Iterable.
// If the underlying Iterable is a Rewinder, Count will rewind the item
// position back to previous state so the Iterable is not consumed (or can
// be consumed again immeidately).
//
// Example:
//   it := New(FromStrings([]string{"a,", "b"}))
//   it.Count() => 2
//   it.Count() => 2
//   it.Filter(func(v interface{}) bool {return v.(string) == "a"}).Count() => 1
func (it *Iter) Count() int {
	return it.impl.count()
}

// Nth returns the n'th item from the Iterable.
// If N isn't in the valid iteration scope, nil will be returned.
// If the Iterable is also a Rewinder, then after retrieving
// the Nth item, the Iterable will be rewinded and assumed to be
// reusable immeidately.
func (it *Iter) Nth(n int) interface{} {
	defer func() {
		if ag, ok := it.impl.item.(Rewinder); ok {
			ag.Rewind()
		}
	}()

	it.impl.advanceBy(n)
	v, _ := it.impl.item.Next()

	return v
}

// Each runs a function against each item for an Iterable
// without changing the item state.
// If the Iterable is also a Rewinder, then after iterating
// all items, the Iterable will be rewinded and assumed to be
// reusable immeidately.
func (it *Iter) Each(f EachFunc) {
	it.impl.each(f)
}

// Into converts self Iterable with underlying type T to another
// Iterable with underlying type U.
// If other is a Resetter, then Reset will be called before the
// conversion, otherwise assume other is clean.
func (it *Iter) Into(target Iterable, as ConvertFunc) *Iter {
	return newFromImpl(it.impl.into(target, as))
}

// From converts other Iterable with type U to self with type T.
// If self is a Resetter, then Reset will be called, otherwise,
// assume clean.
func (it *Iter) From(other Iterable, as ConvertFunc) *Iter {
	return newFromImpl(it.impl.from(other, as))
}

// An Iterable for []string, ready to be consume by an Iterator
// such as the Iter.
// This is the only Iterable implementation provided by the API
// since Go hasn't yet had Generics. It would be tedious if not
// impossible to implement all []T. So if there is a need for
// some T, client will have to implement on thir own.

// IterStrings implements Iterable API for []string.
// IterStrings itself is not thread-safe.
type IterStrings struct {
	idx  int
	data []string
	size int
}

// NewIterStrings creates a new empty IterStrings struct.
func NewIterStrings() *IterStrings {
	return &IterStrings{idx: -1}
}

// FromStrings creates a new IterStrings from a []string.
func FromStrings(s []string) *IterStrings {
	return &IterStrings{idx: -1, data: s, size: len(s)}
}

// New constructs a new empty IterStrings from itself.
func (is *IterStrings) New() (Iterable, error) {
	return NewIterStrings(), nil
}

// Next returns the next string as an interface{}.
// bool indicate whether there is any more to go. If false,
// then IterStrings is exhausted.
func (is *IterStrings) Next() (interface{}, bool) {
	is.idx++
	if is.idx < is.size {
		return is.data[is.idx], true
	}
	return nil, false
}

// Rewind for IterStrings will set the Iterable to its initial
// traversal state and ready for start from beginning again.
func (is *IterStrings) Rewind() {
	is.idx = -1
}

// Reset sets this IterStrings to it's initial state.
// Whatever data hosted would be lost after this call.
func (is *IterStrings) Reset() {
	is.Rewind()
	is.data = nil
	is.size = 0
}

// Add inserts an string as an interface into the IterStrings struct.
func (is *IterStrings) Add(obj interface{}) {
	input := obj.(string)
	is.data = append(is.data, input)
	is.size++
}

// Enumerate returns a pair of {index, string as interface}
// as well as a bool to indicate whether there is more to go.
func (is *IterStrings) Enumerate() (int, interface{}, bool) {
	is.idx++
	if is.idx < is.size {
		return is.idx, is.data[is.idx], true
	}
	return -1, nil, false
}

// String implements the Stringer interface for IterStrings.
func (is *IterStrings) String() string {
	return fmt.Sprintf("%+v", is.data)
}
