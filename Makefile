LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X "main.BuildDate=$(shell /usr/bin/date '+%D %H:%M')"
LDFLAGS += -X "main.BuildCommit=$(shell /usr/bin/git rev-list -1 HEAD)"
LDFLAGS += -extldflags "-static"
# CGO not used in this project, but leave this for future reference
LDFLAGS += -linkmode "external"

GOFLAGS = -tags "netgo,osusergo" -ldflags='$(LDFLAGS)'

##########
# misc
##########

# Update go.mod
update-deps:
	cd ./server/cmd/ && go mod tidy && go get -u
	cd ./scanner/ && go mod tidy && go get -u

release-dirs:
	if [ ! -d "./release" ];			then mkdir ./release			; fi
	if [ ! -d "./release/server" ];	then mkdir ./release/server		; fi
	if [ ! -d "./release/scanner" ];	then mkdir ./release/scanner	; fi


##########
# server
##########

# Delete server release files
server-clean:
	@if [ -e "./release/server/columbus-server" ];			then rm -rf "./release/server/columbus-server"			; fi
	@if [ -e "./release/server/server.conf" ];     			then rm -rf "./release/server/server.conf"				; fi
	@if [ -e "./release/server/columbus-server.service" ];	then rm -rf "./release/server/columbus-server.service"	; fi
	@if [ -e "./release/server/checksum.txt" ];				then rm -rf "./release/server/checksum.txt"				; fi

# Prod build of the server
server-build: release-dirs server-clean
	go build -o release/server/columbus-server $(GOFLAGS) ./server/cmd/.

# Dev build of the server, use --race flag and build onto ./internal directory
server-build-dev: release-dirs
	go build --race -o internal/columbus-server ./server/cmd/.

# Release: build, copy the misc files and create a signed checksum file
server-release: server-clean server-build
	@cp ./server/server.conf ./release/server/server.conf
	@cp ./server/columbus-server.service ./release/server/columbus-server.service
	@cd ./release/server/ && sha512sum * | gpg --local-user daniel@elmasy.com -o checksum.txt --clearsign

##########
# scanner
##########

# Delete scanner release files
scanner-clean:
	@if [ -e "./release/scanner/columbus-scanner" ];			then rm -rf "./release/scanner/columbus-scanner"			; fi
	@if [ -e "./release/scanner/scanner.conf" ];     			then rm -rf "./release/scanner/scanner.conf"				; fi
	@if [ -e "./release/scanner/columbus-scanner.service" ];	then rm -rf "./release/scanner/columbus-scanner.service"	; fi
	@if [ -e "./release/scanner/checksum.txt" ];				then rm -rf "./release/scanner/checksum.txt"				; fi

# Prod build of the scanner
scanner-build: release-dirs scanner-clean
	go build -o release/scanner/columbus-scanner $(GOFLAGS) ./scanner/.

# Dev build of the scanner, use --race flag and build onto ./internal directory
scanner-build-dev: release-dirs
	go build --race -o internal/columbus-scanner ./scanner/.

# Release: build, copy the misc files and create a signed checksum file
scanner-release: scanner-clean scanner-build
	@cp ./scanner/scanner.conf ./release/scanner/scanner.conf
	@cp ./scanner/columbus-scanner.service ./release/scanner/columbus-scanner.service
	@cd ./release/scanner/ && sha512sum * | gpg --local-user daniel@elmasy.com -o checksum.txt --clearsign

##########
# all
##########

release: server-release scanner-release