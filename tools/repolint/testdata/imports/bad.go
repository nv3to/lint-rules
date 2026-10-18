package imports

import (
	"net/http"
	"os/exec"

	_ "example.com/mod/internal/secret"
	alias "example.com/modern/lib"

	"gopkg.in/yaml.v3"
)

import "C"
