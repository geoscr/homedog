SHELL := /bin/bash

# This file is designed to be use with bash tab-completion for Makefile targets
# (try typing "make run-" followed by the tab key twice)

# Import environment
import:
	date
	./env-import.sh

#
# Building (`make all` is easy)
#

homedog: import
	GOPATH=`pwd` go install homedog
