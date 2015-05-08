default: test

.PHONY: test scm git

test:
	go test -v

scm: git

git:
	cd scm && docker build -t bazooka/e2e-git -f Dockerfile-git .
