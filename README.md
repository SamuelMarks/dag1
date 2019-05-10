# DAG1

Novel aBFT consensus

  
# Features
- [x] k-node selection
- [x] k-parent EventBlock
- [x] EventBlock merge
- [ ] DAG1 consensus
    - [x] Dominators
    - [x] Self Dominators
    - [x] Atropos
    - [x] Clotho
    - [x] Frame
    - [x] Frame Received
    - [x] Dominated
    - [x] Lamport Timestamp
    - [x] Atropos Consensus Time
    - [x] Consensus Timestamp
    - [x] Ordering on same Consensus Timestamp (Lamport Timestamp)
    - [ ] Ordering on same Lamport Timestamp (Flag Table)
    - [x] Ordering on same Flag Table (Signature XOR)
    - [x] Transaction submit
    - [x] Consensus Transaction output
    - [ ] Dynamic participants
        - [ ] Peer add
        - [ ] Peer Remove
- [x] Caching for performances
- [x] Sync
- [x] Event Signature
- [ ] Transaction validation
- [ ] Optimum Network pruning

## Dev

### Docker

Create an 3 node dag1 cluster with:

    n=3 BUILD_DIR="$PWD" ./scripts/docker/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [Bash](https://www.gnu.org/software/bash)
  - [git](https://git-scm.com)
  - [Go](https://golang.org)
  - [Glide](https://glide.sh)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
  - [protocol buffers 3](https://github.com/protocolbuffers/protobuf), with: installation of [a release]([here](https://github.com/protocolbuffers/protobuf/releases)) & `go get -u github.com/golang/protobuf/protoc-gen-go`

### Protobuffer 3

This project uses protobuffer 3 for the communication between posets.
To use it, you have to install both `protoc` and the plugin for go code
generation.

Once the stack is setup, you can compile the proto messages by
running this command:

```bash
make proto
```

### DAG1 and dependencies
Clone the [repository](https://github.com/SamuelMarks/dag1) in the appropriate
GOPATH subdirectory:

```bash
$ d="$GOPATH/src/github.com/Fantom-foundation"
$ mkdir -p "$d"
$ git clone https://github.com/SamuelMarks/dag1.git "$d"
```
DAG1 uses [Glide](http://github.com/Masterminds/glide) to manage dependencies.

```bash
$ curl https://glide.sh/get | sh
$ cd "$GOPATH/src/github.com/Fantom-foundation" && glide install
```
This will download all dependencies and put them in the **vendor** folder.

### Other requirements

Bash scripts used in this project assume the use of GNU versions of coreutils.
Please ensure you have GNU versions of these programs installed:-

example for macos:
```
# --with-default-names makes the `sed` and `awk` commands default to gnu sed and gnu awk respectively.
brew install gnu-sed gawk --with-default-names
```

### Testing

DAG1 has extensive unit-testing. Use the Go tool to run tests:
```bash
[...]/dag1$ make test
```

If everything goes well, it should output something along these lines:
```
?   	github.com/SamuelMarks/dag1/cmd/dummy	[no test files]
?   	github.com/SamuelMarks/dag1/cmd/dummy/commands	[no test files]
?   	github.com/SamuelMarks/dag1/cmd/dummy_client	[no test files]
?   	github.com/SamuelMarks/dag1/cmd/dag1	[no test files]
?   	github.com/SamuelMarks/dag1/cmd/dag1/commands	[no test files]
?   	github.com/SamuelMarks/dag1/tester	[no test files]
ok  	github.com/SamuelMarks/dag1/src/common	(cached)
ok  	github.com/SamuelMarks/dag1/src/crypto	(cached)
ok  	github.com/SamuelMarks/dag1/src/difftool	(cached)
ok  	github.com/SamuelMarks/dag1/src/dummy	0.522s
?   	github.com/SamuelMarks/dag1/src/dag1	[no test files]
?   	github.com/SamuelMarks/dag1/src/log	[no test files]
?   	github.com/SamuelMarks/dag1/src/mobile	[no test files]
ok  	github.com/SamuelMarks/dag1/src/net	(cached)
ok  	github.com/SamuelMarks/dag1/src/node	9.832s
?   	github.com/SamuelMarks/dag1/src/pb	[no test files]
ok  	github.com/SamuelMarks/dag1/src/peers	(cached)
ok  	github.com/SamuelMarks/dag1/src/poset	9.627s
ok  	github.com/SamuelMarks/dag1/src/proxy	1.019s
?   	github.com/SamuelMarks/dag1/src/proxy/internal	[no test files]
?   	github.com/SamuelMarks/dag1/src/proxy/proto	[no test files]
?   	github.com/SamuelMarks/dag1/src/service	[no test files]
?   	github.com/SamuelMarks/dag1/src/utils	[no test files]
?   	github.com/SamuelMarks/dag1/src/version	[no test files]
```

## Cross-build from source

The easiest way to build binaries is to do so in a hermetic Docker container.
Use this simple command:

```bash
[...]/dag1$ make dist
```
This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

```bash
[...]/dag1$ tree --charset=nwildner build
build
|-- dist
|   |-- dag1_0.4.3_SHA256SUMS
|   |-- dag1_0.4.3_darwin_386.zip
|   |-- dag1_0.4.3_darwin_amd64.zip
|   |-- dag1_0.4.3_freebsd_386.zip
|   |-- dag1_0.4.3_freebsd_arm.zip
|   |-- dag1_0.4.3_linux_386.zip
|   |-- dag1_0.4.3_linux_amd64.zip
|   |-- dag1_0.4.3_linux_arm.zip
|   |-- dag1_0.4.3_windows_386.zip
|   `-- dag1_0.4.3_windows_amd64.zip
|-- dag1
`-- pkg
    |-- darwin_386
    |   `-- dag1
    |-- darwin_386.zip
    |-- darwin_amd64
    |   `-- dag1
    |-- darwin_amd64.zip
    |-- freebsd_386
    |   `-- dag1
    |-- freebsd_386.zip
    |-- freebsd_arm
    |   `-- dag1
    |-- freebsd_arm.zip
    |-- linux_386
    |   `-- dag1
    |-- linux_386.zip
    |-- linux_amd64
    |   `-- dag1
    |-- linux_amd64.zip
    |-- linux_arm
    |   `-- dag1
    |-- linux_arm.zip
    |-- windows_386
    |   `-- dag1.exe
    |-- windows_386.zip
    |-- windows_amd64
    |   `-- dag1.exe
    `-- windows_amd64.zip

11 directories, 29 files
```
