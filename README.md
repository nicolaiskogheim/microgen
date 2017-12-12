# Microgen

Tool to generate microservices from service interfaces, using
[go-kit](https://gokit.io/). The goal is to be able to generate boilerplate,
and keep that boilerplate in sync with changes in the service.

# About this fork

This is a fork of [devimteam/microgen](https://github.com/devimteam/microgen).
I wanted to experiment a little to see if I could improve on it, but I also
want to change how microgen lays out files, how it names methods and possibly
other things. I will try and structure my work such that it can be
cherry-picked upstream.

## Roadmap

#### Directory and file structure change

I want the default layout to mimick the examples in go-kit, like the [addservice](https://github.com/go-kit/kit/tree/master/examples/addsvc):

```
addsvc
├── cmd
│   ├── addcli
│   │   └── addcli.go
│   └── addsvc
│       ├── addsvc.go
│       ├── pact_test.go
│       └── wiring_test.go
├── pb
│   ├── addsvc.pb.go
│   ├── addsvc.proto
│   └── compile.sh
├── pkg
│   ├── addendpoint
│   │   ├── middleware.go
│   │   └── set.go
│   ├── addservice
│   │   ├── middleware.go
│   │   └── service.go
│   └── addtransport
│       ├── grpc.go
│       ├── http.go
│       └── thrift.go
├── README.md
└── thrift
    ├── addsvc.thrift
    ├── compile.sh
    └── gen-go
        └── addsvc
            ├── add_service-remote
            │   └── add_service-remote.go
            ├── addsvc-consts.go
            ├── addsvc.go
            └── GoUnusedProtection__.go
```

#### Other stuff

After the directory/file structure stuff is done I'll have a look at other stuff.

## Install

```
go get -u github.com/devimteam/microgen/cmd/microgen
```

> If you have problems with building microgen, install
> [dep](https://github.com/golang/dep) and use `dep ensure` to install correct
> versions of dependencies
> ([#29](https://github.com/devimteam/microgen/issues/29)).

## Usage

``` sh
microgen [OPTIONS]
```

`microgen` will look for the first `type * interface` definition in the
provided file, that contains `// @microgen` in the interface docs.

Generation parameters is provided through ["tags"](#tags) in interface docs
after the `// @microgen` tag (space before is @ __required__).

### Options

| Name   | Default    | Description                                                                   |
|:------ |:-----------|:------------------------------------------------------------------------------|
| -file  | service.go | Relative path to source file with service interface                           |
| -out   | .          | Relative or absolute path to directory, where you want to see generated files |
| -force | false      | With flag generate stub methods.                                              |
| -help  | false      | Print usage information                                                       |

### Markers

Markers is general tags that affect generation.
The syntax is: `// @<tag-name>:`

#### @microgen

Main tag for `microgen`. Microgen scans the input file for the first interface
in which docs contains this tag. Add [tags](#tags), separated by comma after
`@microgen` to generate code for it:

```go
// @microgen middleware, logging
type StringService interface {
    ServiceMethod()
}
```

#### @protobuf

Specify which protobuf implementation to use with `grpc`.
**Required for `grpc`, `grpc-server`, `grpc-client` generation.**

```go
// @microgen grpc-server
// @protobuf github.com/user/repo/path/to/protobuf
type StringService interface {
    ServiceMethod()
}
```

#### @grpc-addr

gRPC address tag is used for gRPC go-kit-based client generation.
**Required for `grpc-client` generation.**

```go
// @microgen grpc
// @protobuf github.com/user/repo/path/to/protobuf
// @grpc-addr some.service.address
type StringService interface {
    ServiceMethod()
}
```

#### @force

Files generated for the `middleware` and `logging` tags are overwritten each
time. Files generated by `http` and `grpc`, for instance, will not be
overwritten by default. Use (the) `@force` to overwrite all files.


### Method's tags

#### @logs-ignore

Tells logging middleware which method arguments and/or return values shouldn't
be logged, like passwords or files. The first argument, `ctx contex.Context`
is ignored by default.

```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

#### @logs-len

Instructs logging middleware to log `len(arg)`. `arg` is still logged unless
`@logs-ignore arg` is specified.

```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    // @logs-len data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

> Without `@logs-ignore data` in the example above, we would log both `data`
> and `len(data)`

### Tags

| Tag         | Description                                                           | Overwrites existing files |
|:------------|:----------------------------------------------------------------------|---------------------------|
| middleware  | General application middleware interface.                                                   | Yes |
| logging     | Middleware that writes to logger all request/response information with handled time.        | Yes |
| recover     | Middleware that recovers panics and writes errors to logger.                                | Yes |
| grpc-client | Generates client for grpc transport with request/response encoders/decoders.                | No  |
| grpc-server | Generates server for grpc transport with request/response encoders/decoders.                | No  |
| grpc        | Generates client and server for grpc transport with request/response encoders/decoders.     | No  |
| http-client | Generates client for http transport with request/response encoders/decoders.                | No  |
| http-server | Generates server for http transport with request/response encoders/decoders.                | No  |
| http        | Generates client and server for http transport with request/response encoders/decoders.     | No  |
| main        | Generates basic `package main` for starting service. Affected by other tags                 | No  |

> Use the `@force` tag, or the `-force` flag to overwrite all files.

### Files

| Name  | Default path |  Generation logic |
|-----------------------|----------------------------|-----|
| Service interface     | ./service.go               | Add service entity, constructor, and methods if missing.|
| Exchanges             | ./exchanges.go             | Overwrites old file every time.|
| Endpoints             | ./endpoints.go             | Overwrites old file every time.|
| Middleware            | ./middleware/middleware.go | Overwrites old file every time.|
| Logging middleware    | ./middleware/logging.go    | Overwrites old file every time.|
| Recovering middleware | ./middleware/recovering.go | Overwrites old file every time.|

## Example

Follow this short guide to try the `microgen` tool.

1. Create file `service.go` inside GOPATH and add code below.

```go
package stringsvc

import (
    "context"
)

// @microgen http, grpc, middleware, logging, recover, main
// @protobuf github.com/devimteam/proto-utils
// @grpc-addr test.address
type StringService interface {
    Uppercase(ctx context.Context, str string) (ans string, err error)
    Count(ctx context.Context, text string, symbol string) (count int, positions []int)
}
```

2. Open command line next to your `service.go`.
3. Enter `microgen`. __*__
4. You should see something like this:

```
@microgen 0.5.0
Tags: middleware, logging, grpc, http, recover, main
New cmd/string_service/main.go
Add .../svc.go
New exchanges.go
New endpoints.go
New middleware/middleware.go
Add transport/converter/http/exchange_converters.go
New middleware/logging.go
New middleware/recovering.go
New transport/grpc/server.go
New transport/grpc/client.go
New transport/converter/protobuf/type_converters.go
New transport/converter/protobuf/endpoint_converters.go
All files successfully generated
```

5. Now, add and generate protobuf file (if you use grpc transport) and write transport converters (from protobuf/json to golang and _vise versa_).
6. Use endpoints in your `package main` or wherever you want. (tag `main` generates some code for `package main`)

__*__ `GOPATH/bin` should be in your PATH.

## Interface declaration rules

Generation may fail if these aren't adhered to.

* All interface method's arguments and results should be named and should be different.
* First argument of each method should be of type `context.Context` (from [standard library](https://golang.org/pkg/context/)).
* Last result should be builtin `error` type.
---
* Name of _protobuf_ service should be the same as the interface name.
* Function names in _protobuf_ should be the same as in interface.
* Message names in _protobuf_ should be named `<FunctionName>Request` and `<FunctionName>Response` for request and response messages, respectively.
* Field names in _protobuf_ messages should be the same as in interface methods (_protobuf_ - snake_case, interface - camelCase).

## Dependencies

After generation your service may depend on this packages:

```
    "net/http"      // http
    "bytes"
    "encoding/json" // http
    "io/ioutil"
    "strings"
    "net"           // http and grpc listeners
    "net/url"       // http
    "fmt"
    "context"
    "time"          // logging
    "os"            // for signal handling and os.Stdout
    "os/signal"     // for signal handling
    "syscall"       // for signal handling
    "errors"        // for error handling

    "google.golang.org/grpc"                    // grpc
    "golang.org/x/net/context"
    "github.com/go-kit/kit"                     // grpc
    "github.com/golang/protobuf/ptypes/empty"   // grpc
```
