package main

import (
	"alpha.dagger.io/git"
)

manifest: git.#Repository & {
	remote: "https://github.com/dagger/examples.git"
	ref:    "main"
	subdir: "todoapp/k8s"
}

repository: git.#Repository & {
	remote: "https://github.com/dagger/examples.git"
	ref:    "main"
	subdir: "todoapp"
}

registry: "gcr.io/dagger-ci/test"
