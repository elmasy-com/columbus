LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X "main.BuildDate=$(shell /usr/bin/date '+%D %H:%M')"
LDFLAGS += -X "main.BuildCommit=$(shell /usr/bin/git rev-list -1 HEAD)"
LDFLAGS += -extldflags "-static"
# CGO not used in this project, but leave this for future reference
LDFLAGS += -linkmode "external"

GOFLAGS = -tags "netgo,osusergo" -ldflags='$(LDFLAGS)'

# Update go.mod
update-deps:
	cd ./server/cmd/ && go mod tidy && go get -u

# Delete server release files
server-clean:
	@if [ -e "./release/server/columbus-server" ];			then rm -rf "./release/server/columbus-server"			; fi
	@if [ -e "./release/server/server.conf" ];     			then rm -rf "./release/server/server.conf"				; fi
	@if [ -e "./release/server/columbus-server.service" ];	then rm -rf "./release/server/columbus-server.service"	; fi

# Prod build of the server
server-build: server-clean
	go build -o release/server/columbus-server $(GOFLAGS) ./server/cmd/.

# Dev build of the server, use --race flag and build onto ./internal directory
server-build-dev:
	go build --race -o internal/columbus-server ./server/cmd/.

# Release: build, copy the misc files and create a signed checksum file
server-release: server-clean server-build
	@cp ./server/server.conf ./release/server/server.conf
	@cp ./server/columbus-server.service ./release/server/columbus-server.service
	@cd ./release/server/ && sha512sum * | gpg --local-user daniel@elmasy.com -o checksum.txt --clearsign