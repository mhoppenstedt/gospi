help:
	@echo "You can perform the following:"
	@echo ""
	@echo "  check         Format, lint, vet, and test Go code"
	@echo "  generate      Run \`go generate\`"
	@echo "  local         Build for local development OS"
	@echo "  arm           Build for ARM architecture"
	@echo "  linux         Build for Linux OS"
	@echo "  debug         Build for local development OS and run the application"
	@echo "  release       Build a zipped bundle ready used for the web updater on a Linux target"

# Remove all of the develoment images
clean:
	go clean
	-rm -r build/*

# Format, lint, vet, and test the Go code
check:
	@echo 'Formatting, linting, vetting, and testing Go code'
	go fmt ./...
	golint ./...
	go vet ./...
	go test ./...

#  Calculate the code coverage
coverage:
	go test -coverprofile=build/coverage.out && go tool cover -html=build/coverage.out
	
#  Run `go generate`
generate:
	go generate

#  Compile the project to run locally on your machine
local: generate
	go build -o build/battcs

#  Compile the project to run on an ARM processor 
arm: generate
	GOARCH=arm GOOS=linux go build -o build/battcs-arm

#  Compile the project to run on a standard linux OS
linux: generate
	GOOS=linux go build -o build/battcs-linux

#  Build and run a local build
debug: clean local
	if ! [ -a debug.json ]; then cp sys/opt/battcs/applianceConfig.json debug.json; echo "Created debug configuration"; fi;

	build/battcs debug.json

#  Build a release for a linux environment
release: clean arm
	echo "=========================================================="
	echo "=                                                        ="
	echo "= Did you update the version numbers in the DEB and SW?? ="
	echo "=                                                        ="
	echo "=========================================================="
	mkdir -p build/deb/opt/battcs/
	cp build/battcs-arm build/deb/opt/battcs/battcsAppliance
	cp -r sys/DEBIAN build/deb/
	dpkg -b build/deb/ build/battcsAppliance_1.0.0_arm.deb

#  Build the tar file used to provision a new server
config:
	cd sys && tar --exclude='.DS_Store' --exclude='DEBIAN' -cvf ../build/battcsAppliance.tar opt etc
