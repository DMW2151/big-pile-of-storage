#
# Makefile that managed the build, testing, accceptance testing, and release of the
# CORTX Terraform provider.
#

## Args

# Build Parameters
HOSTNAME=dmw2151.com
NAMESPACE=terraform
NAME=cortx
BINARY=terraform-provider-${NAME}

# NOTE: Get OSType (e.g. darwin, linux, freebsd, windows, etc...) and ArchType (e.g. amd64, 
# armv7, etc...). $OS_ARCH should be something like darwin_amd64. 
#
# TODO/WARN: Expect $TARGET_ARCH to match an entry in `go tool dist list`, no clean way to get 
# those names so hardcoded for the moment!

TARGET_OS=darwin# Consider: ${OSTYPE//[0-9.]/}
TARGET_ARCH=amd64  # Consider: uname -r

OS_ARCH=${TARGET_OS}_${TARGET_ARCH}

# NOTE: `version` tag needed for Terraform provider definition, hardcoded to 0.0.1a 
# for CORTX Hackathon - can auto-increment on release to tagged branch, etc...
VERSION=0.0.1

# NOTE: Testing Arguments - Manually input `$TESTARGS` if needed, bare `go test` is fine
# as of 0.0.1a.
TEST?=$$(go list ./... | grep -v 'vendor')
TESTARGS=""


## Stages

default: install

install: 
	env GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -o ${BINARY}
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
