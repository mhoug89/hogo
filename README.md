# HoGo

HoGo is a set of core Go libraries containing types that I consistently find
myself wishing Go provided in its standard library.

I chose to include multiple libraries in one module to minimize the amount of
dependency management I need to do when starting a new project in Go, i.e. I can
run a single `go get` command:

```bash
go get github.com/mhoug89/hogo
```
