# noname

A tree-walking interpreter written in Go.

## Running

```bash
go run main.go
```

This starts an interactive REPL.

## Language Reference

### Types

| Type     | Examples                        |
|----------|---------------------------------|
| Integer  | `5`, `42`, `-7`                 |
| Boolean  | `true`, `false`                 |
| String   | `"hello"`, `"world"`            |
| Null     | returned by empty `if` with no match |
| Function | `fn(x) { x + 1 }`              |

---

### Variables

```
let name = value;
```

```
let x = 10;
let greeting = "hello";
```

Variables are immutable once bound. Re-binding with `let` creates a new binding in the current scope.

---

### Arithmetic

```
5 + 3    // 8
10 - 4   // 6
3 * 7    // 21
20 / 4   // 5
```

Operator precedence follows standard math rules. Use parentheses to group:

```
(2 + 3) * 4   // 20
2 + 3 * 4     // 14
```

---

### Comparison & Boolean Operators

```
5 == 5    // true
5 != 3    // true
5 > 3     // true
5 < 3     // false

true == true    // true
true != false   // true

!true     // false
!false    // true
!!true    // true
```

---

### If / Else

`if` is an expression — it returns a value.

```
if (condition) { consequence } else { alternative }
```

```
let x = if (10 > 5) { 10 } else { 5 };
x   // 10
```

The `else` branch is optional. If the condition is false and there is no `else`, the expression returns `null`.

```
if (false) { 1 }   // null
```

---

### Functions

Functions are first-class values defined with `fn`.

```
fn(param1, param2) { body }
```

```
let add = fn(x, y) { x + y };
add(3, 4)   // 7
```

The last evaluated expression in the body is the return value. You can also return early with `return`:

```
let max = fn(a, b) {
    if (a > b) { return a; };
    b
};

max(3, 7)    // 7
max(10, 2)   // 10
```

#### Higher-order functions

Functions can take other functions as arguments:

```
let apply = fn(f, x) { f(x) };
let double = fn(x) { x * 2 };

apply(double, 5)   // 10
```

#### Closures

Functions close over their enclosing scope:

```
let makeAdder = fn(x) {
    fn(y) { x + y }
};

let addFive = makeAdder(5);
addFive(3)    // 8
addFive(10)   // 15
```

---

### Strings

String literals are enclosed in double quotes.

```
let name = "noname";
```

Concatenate with `+`:

```
let hello = "hello" + ", " + "world";
hello   // hello, world
```

Compare with `==` and `!=`:

```
"foo" == "foo"   // true
"foo" != "bar"   // true
```

---

### Builtin Functions

| Function   | Description                          | Example                        |
|------------|--------------------------------------|--------------------------------|
| `puts(x)`  | Print value(s), returns null         | `puts("hello")`                |
| `len(s)`   | Length of a string                   | `len("hello")` → `5`          |
| `str(x)`   | Convert any value to string          | `str(42)` → `"42"`            |
| `type(x)`  | Return the type name as a string     | `type(42)` → `"INTEGER"`      |

```
puts("hello, " + "world")
// hello, world

len("noname")
// 6

str(100)
// 100

let age = 30;
puts("age is: " + str(age))
// age is: 30

type(42)       // INTEGER
type("hello")  // STRING
type(true)     // BOOLEAN
type(fn(x){x}) // FUNCTION
```

---

### Return

`return` exits the current function with a value:

```
let sign = fn(n) {
    if (n > 0) { return 1; };
    if (n < 0) { return -1; };
    0
};

sign(42)    // 1
sign(-7)    // -1
sign(0)     // 0
```

---

### Errors

The interpreter reports errors at runtime:

```
foobar              // ERROR: identifier not found: foobar
5 + true            // ERROR: type mismatch: INTEGER + BOOLEAN
let f = 5; f(10)    // ERROR: not a function: INTEGER
```

---

## Project Structure

```
.
├── ast/         AST node definitions
├── evaluator/   Tree-walking evaluator
├── lexer/       Tokenizer
├── object/      Runtime value types and environment
├── parser/      Pratt parser
├── repl/        Interactive REPL
└── token/       Token type definitions
```

## Running Tests

```bash
go test ./...
```
