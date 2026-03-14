# [FIXED in 0.15.0] Match Expression Type Inference Fails for Go Interop Types and Block Branches

## Summary

The GALA transpiler fails with `[SemanticError] cannot infer result type of match expression: no branch returns a concrete type` when using `match` as an expression in several common patterns.

## Environment

- GALA version: 0.14.0
- OS: Windows 11

## Reproduction Cases

### Case 1: Match returning Go interop types

Match on a string where branches return Go types (e.g., `*exec.Cmd`):

```gala
package main

import (
    "os/exec"
    "runtime"
)

func main() {
    val goos = runtime.GOOS
    val cmd = goos match {
        case "windows" => exec.Command("notepad")
        case "darwin" => exec.Command("open", ".")
        case _ => exec.Command("xdg-open", ".")
    }
    _ = cmd.Start()
}
```

**Expected**: All branches return `*exec.Cmd`, type should be inferred.
**Actual**: `cannot infer result type of match expression`

**Workaround**: Replace with if-else chain.

### Case 2: Match on Option with block branch

Match on `Option[string]` where one branch is a simple expression and the other is a block:

```gala
package main

import (
    . "martianoff/gala/collection_immutable"
    "os"
    "os/exec"
)

func findBinary() string {
    val candidates = ArrayOf("/usr/bin/gala", "/usr/local/bin/gala")
    val found = candidates.Find((c) => {
        val _, err = os.Stat(c)
        return err == nil
    })

    return found match {
        case Some(path) => path
        case None() => {
            val p, err = exec.LookPath("gala")
            if err == nil {
                return p
            }
            return ""
        }
    }
}

func main() {
    Println(findBinary())
}
```

**Expected**: Both branches return `string`, type should be inferred.
**Actual**: `cannot infer result type of match expression`

**Workaround**: Replace with `if found.IsDefined() { return found.Get() }` followed by imperative code.

### Case 3: Match on Try inside FoldLeft lambda

Match on `Try[string]` where one branch is a block and the other is a simple expression:

```gala
package main

import . "martianoff/gala/collection_immutable"

embed val data EmbeddedFS = "examples/*"

func process() string {
    val names = ArrayOf("hello", "world")
    val results = names.FoldLeft(EmptyArray[string](), (acc, name) => {
        return data.ReadString(s"examples/$name.txt") match {
            case Success(content) => {
                val upper = content
                return acc.Append(upper)
            }
            case Failure(_) => acc
        }
    })
    return results.Head()
}

func main() {
    Println(process())
}
```

**Expected**: Both branches return `Array[string]`, type should be inferred.
**Actual**: `cannot infer result type of match expression`

**Workaround**: Replace match with `if result.IsSuccess() { ... } else { ... }`.

### Case 4: Match on string for val assignment

Simple string-to-string match used to assign a `val`:

```gala
package main

import "os"

func main() {
    val envPort = os.Getenv("PORT")
    val port = envPort match {
        case "" => "3000"
        case _ => envPort
    }
    Println(port)
}
```

**Expected**: Both branches return `string`, trivially inferable.
**Actual**: `cannot infer result type of match expression`

**Workaround**: Use `var port = "3000"; if os.Getenv("PORT") != "" { port = os.Getenv("PORT") }`

## Common Pattern

All cases share the characteristic that match is used as an **expression** (assigned to `val` or used in `return`). The transpiler cannot determine the result type even when all branches clearly return the same type.

## Suggested Fix

When a match is used as an expression:
1. Infer the type from the first branch that returns a concrete expression
2. Verify all other branches return a compatible type
3. For Go interop types, resolve the Go return type of called functions (e.g., `exec.Command` -> `*exec.Cmd`)
