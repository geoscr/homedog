SHELL := /bin/bash

# Import environment
import:
	date
	./env-import.sh

#
# Building
#

ensure: import
	cd src/homedog  ; \
	dep ensure      ; \
	cd ../..

build: import
	GOPATH=`pwd` go install homedog

run: import
	HOMEDOG_CONFIG=../../config/config.json ./bin/homedog
