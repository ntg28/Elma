# Elma

Elma is a transpiler from go to javascript.

### NOTES

* This is a one week project so it is not finished and its not
in development (for now...)

* The library is short but there are few examples of how to bind
go functions to javascript

* I implemented a simple pomodoro app using `Elma` is located at
`exs/pomodoro` I think it is a good source for getting the idae of
how to use `Elma`
```bash
go run . ./exs/pomodoro
```

### IMPORTANT

* *DO NOT* create functions with the same name as library functions
> if the code find a function it tries to look on all the functions
> and try to expand to the js binding.

* The order of the values in struct construction *IS IMPORT*
because it transpile to a function call that construct the object
if not respected the order of the parameters will be wrong.
```go
type Person struct {
    name string
    age int
}

// WRONG
a := Person{
    age: 12,
    name: "john",
}

// RIGHT
a := Person{
    name: "john",
    age: 12,
}

// RIGHT
a := Person{
    "john",
    12,
}
```
