# RDT
A reliable data transfer protocol over UDP using GBN
## Compilation
Execute the build script on any machine with golang installed, passing in target OS and architecture.
```bash
./build.sh <os> <arch>
```
*The machine where the program is built can have a different os/arch than the target.

You can compile for the all of the most common targets (os=linux/darwin/windows) and (arch=amd64/arm64)
with
```bash
./buildall.sh
```
## Execution
### Environment
Configure protocol by setting environment variables:
```bash
export PORT=<port> MAX_SEQ_NO=<max sequence no.> WINDOW_SIZE=<gbn window size>
```
A window size of 1 will degenerate the pipelined GBN protocol to the regular non-pipelined RDT protocol
### Server
Execute the following binary:
```bash
./bin/<os>/<arch>/server
```
### Client
Execute the following binary, passing in the hostname of the server
```bash
./bin/<os>/<arch>/client <servername>
```

## Tips
The target for a linux machine running on an intel/amd CPU will be
os=linux and arch=amd64