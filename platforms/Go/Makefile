SHELL := /bin/bash

# Import environment
import:
	date
	./env-import.sh

#
# Building
#

homedog: import
	GOPATH=`pwd` go install homedog
