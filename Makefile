BINARY_NAME=gonalmatrix

build:
	GOARCH=amd64 GOOS=freebsd go build -o ${BINARY_NAME}-freebsd ./cmd/gonalmatrix
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux ./cmd/gonalmatrix

clean:
	go clean
	rm ${BINARY_NAME}-freebsd
	rm ${BINARY_NAME}-linux
