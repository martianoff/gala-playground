# [FIXED in 0.17.0] Try Lambda: `(T, error)` Not Destructured for Chained Method Calls

## Summary

Inside a `Try()` lambda, Go functions returning `(T, error)` are correctly destructured when called as standalone statements, but NOT when the call is a **chained method** like `obj.Method1().Method2()`.

## Environment

- GALA version: 0.16.1
- OS: Windows 11

## Reproduction

```gala
package main

import "os/exec"

func main() {
    // WORKS: standalone call
    val result1 = Try(() => exec.LookPath("gala"))
    Println(result1)

    // FAILS: chained method call
    val result2 = Try(() => exec.Command("gala", "version").Output())
    Println(result2)
}
```

## Generated Go (incorrect)

```go
// Standalone — correctly destructured:
std.Try[string]{}.Apply(func() string {
    _v0, _err := exec.LookPath("gala")
    if _err != nil { panic(_err) }
    return _v0
})

// Chained — NOT destructured:
std.Try[[]byte]{}.Apply(func() []byte {
    return exec.Command("gala", "version").Output()  // returns ([]byte, error)!
})
```

## Error

```
too many return values: have ([]byte, error), want ([]byte)
```

## Expected

```go
std.Try[[]byte]{}.Apply(func() []byte {
    _v0, _err := exec.Command("gala", "version").Output()
    if _err != nil { panic(_err) }
    return _v0
})
```

## Workaround

Break the chain into separate statements:

```gala
func getGalaVersion(galaBin string) string {
    val cmd = exec.Command(galaBin, "version")
    val out, err = cmd.Output()
    if err != nil {
        return "unknown"
    }
    return strings.TrimSpace(string(out))
}
```
