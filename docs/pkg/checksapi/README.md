<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# checksapi

```go
import "github.com/luisdavim/synthetic-checker/pkg/checksapi"
```

## Index

- [func New(chkr *checker.Runner, srvCfg server.Config, opts Options) *server.Server](<#func-new>)
- [type Options](<#type-options>)


## func [New](<https://github.com/luisdavim/synthetic-checker/blob/main/pkg/checksapi/server.go#L17>)

```go
func New(chkr *checker.Runner, srvCfg server.Config, opts Options) *server.Server
```

New creates a new check API server

## type [Options](<https://github.com/luisdavim/synthetic-checker/blob/main/pkg/checksapi/server.go#L10-L14>)

```go
type Options struct {
    FailStatus     int
    DegradedStatus int
    ExposeConfig   bool
}
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
