# goiter v0.1.0

An Iterator API that inspired from Rust's Iterator Trait. [WIP]

A few notes for this API:

🩸 The design of this Iterator API is heavily inspired by the 🦀Rust's Iterator trait. It tries to bring some familiar mindsets b/w Go and Rust. You could think of it as a study series, or design pattern research, or whatever.

🩸 Thus, the work is still In-Progress, whenever I see a good fit for expressing a pattern in Go, I will try to implement it. Or if you'd like to contribute, feel free for PRs. 😉

🩸 At this point, Go still doesn't have Generics, but I think we are [close](https://blog.golang.org/generics-next-step). So there is a reasonable tendency that expect for the `interface{}` based utility functions, we probably shouldn't add too many type specific Iterators. The T Iterator would probably be soon implemented by using the TRUE Iter<T> some time next year. For now, given the prevalence of go string type, I implemented `IterStrings`, it could serve as an example to implement other types if needed. Once Go's Generics is able, we shall be able to rewrite the implementation and all the utility functions as well, plus, it would be much appropriate to write various different kind of converters/adapters b/w different generic types {T, U, etc}.

🩸 Despite its small size, the API is already quite powerful within what it is capable of (see some examples in tests).

🩸 Go's abstraction is not without a cost. This Iterator's performance is not on par with the plain old for loop version ( See benchmarks ). If you absolutely care about performance, you probably should look somewhere else. The good part is it does not have any external dependencies, it uses no reflect APIs of any sort. I reasonably believe that once the Go compiler is able to generate machine code directly from generic type system, the performance shall be a lot better (e.g. all runtime type cast would go away).
