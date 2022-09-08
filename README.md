# saver-grpc-lib

For regenerating proto structures use:

```
protoc ./proto/service.proto --gogofaster_out=GOPATH/src --go-grpc_out=GOPATH/src
```

Proto compiler installation:

```
brew install protoc-gen-gogofaster
```

```
brew install protoc-gen-go-grpc 
```
