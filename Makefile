# Detect Rancher Desktop
ifeq ($(shell command -v nerdctl 2> /dev/null),)
	# Use Docker if nerdctl is not found
	cmd := docker
else
	# Use nerdctl if it is found
	cmd := nerdctl
endif

# local test target
test:
	${cmd} build --target test --rm .

.PHONY: test
