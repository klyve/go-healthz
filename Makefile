test:
	go test -v
cover:
	rm -rf *.coverprofile
	go test -coverprofile=healthz.coverprofile
	gover
	go tool cover -html=healthz.coverprofile