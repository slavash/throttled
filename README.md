#### Throttled

TCP networking listener implementation with rate limit in mind

### Usage
```go
l, err := net.Listen("tcp4", "localhost:8080")
if err != nil {
    fmt.Println(err)
    return
}

limited := throttled.NewListener(l)
limited.SetLimits(serverLimitBytePerSec, connLimitBytePerSec)

defer func() {
    if e := limited.Close(); e != nil {
        fmt.Printf("faield to shotdown the server: %s\n", e)
    }
}()

fmt.Printf("Listening on %s\n", l.Addr())

for {
    c, err := limited.Accept()
    if err != nil {
        fmt.Println(err)
        return
    }

    go handleConnection(c)
}
```