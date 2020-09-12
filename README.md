# LOG file reader

## Usage
Possible to give path of file to analyze and trace level.
Fore more detail, execute "go run . -help" to get instructions on available options.
E.g., "go run . -level=TRACE" to see all the log lines with level set as trace.

k8s package can retrieve logs for all the pods

## Build for windows from mac or linux
env GOOS=windows GOARCH=amd64 build .