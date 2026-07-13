# LinaPro plugin Makefile.

PLUGIN_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
REPO_ROOT := $(abspath $(PLUGIN_ROOT)/../../..)

include $(REPO_ROOT)/hack/makefiles/plugin.codegen.mk
