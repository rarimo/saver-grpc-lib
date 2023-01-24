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

This lib contains the following components:
- Broadcaster connector
- GRPC service declaration
- Voter SDK

## Voter SDK

Use `gitlab.com/rarimo/savers/saver-grpc-lib/voter` package in savers to implement logic of monitoring anv voting operations.

The `voter.Voter` structure implements the common logic of verifying and voting for the certain operation. 
It uses `voter.IVerifier`component to verify operation of the certain types. Mapping with `rarimotypes.OpType => IVerifier` 
correspondence definition should be provided into the `voter.NewVoter` constructor. 

The `gitlab.com/rarimo/savers/saver-grpc-lib/voter/verifiers` package provides common logic of already designed IVerifiers. 

The `verifiers.TransferVerifier` implements logic for verifying transfer operations. All savers should also implement  
`verifiers.ITransferOperator` that injects custom chain logic into `verifiers.TransferVerifier` to verify operation. 

## GRPC

Grpc package defines only one entrypoint `Revote` that should be used to trigger re-verifying and re-voting for the 
certain operation.