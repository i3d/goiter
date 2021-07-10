# goiter v0.1.1

An Iterator API that inspired from Rust's Iterator Trait. [WIP]

A few notes for this API:

🩸 The design of this Iterator API is heavily inspired by the 🦀Rust's Iterator trait. It tries to bring some familiar mindsets b/w Go and Rust. You could think of it as a study series, or design pattern research, or whatever.

🩸 Thus, the work is still In-Progress, whenever I see a good fit for expressing a pattern in Go, I will try to implement it. Or if you'd like to contribute, feel free for PRs. 😉

🩸 At this point, Go still doesn't have Generics, but I think we are [close](https://blog.golang.org/generics-next-step). So there is a reasonable tendency that expect for the `interface{}` based utility functions, we probably shouldn't add too many type specific Iterators. The T Iterator would probably be soon implemented by using the TRUE Iter<T> some time next year. For now, given the prevalence of go string type, I implemented `IterStrings`, it could serve as an example to implement other types if needed. Once Go's Generics is able, we shall be able to rewrite the implementation and all the utility functions as well, plus, it would be much appropriate to write various different kind of converters/adapters b/w different generic types {T, U, etc}.

🩸 Despite its small size, the API is already quite powerful within what it is capable of (see some examples in tests).

🩸 This Iterator implementation is not thread-safe. This is generally true for all Iterator implementations from most languages.

🩸 Go's abstraction is not without a cost. This Iterator's performance is not on par with the plain old for loop version ( See benchmarks ). If you absolutely care about performance, you probably should look somewhere else. The good part is it does not have any external dependencies, it uses no reflect APIs of any sort. I reasonably believe that once the Go compiler is able to generate machine code directly from generic type system, the performance shall be a lot better (e.g. all runtime type cast would go away).

🩸 This package's Iter utilitiy functions are not lazy, in other words, not like Rust's behavior where only consuming APIs (e.g. collect<T>) will materialize the collection from the Iterator, it materialize the Iterator immeidately upon calling. In most cases, it will produce a new Iter instead of in-line mutating the existing one.

🩸 Read-only functions are Rewindable, meaning, if an Iterator implements `Rewinder`, as soon as the read is done, the Iterator is rewinded to it's previous state (or whatever state the `Rewinder` is defined) and assumed to be ready to consume again.
For example, you could call `iter.Nth(10)` and then immeidately `iter.Nth(5)` without any problem if `iter` is also a `Rewinder`. This is different than the Rust version of read-only consume functions where once consumed, the Iterator is no longer available.

🩸 Go does not have Enum objects natively and it probably not needed to build such abstraction. While Rust's embrassing `Option<T>` and `Result<T,E>` a lot in its stdlib, this implementation will just stick with Go's multi-return pattern. There is nothing wrong with returning `(nil, !more) ` indicating there is no more to go. This also handles `nil` element correctly.

### TODO: add examples in go doc.
