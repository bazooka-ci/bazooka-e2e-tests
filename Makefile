default: test

.PHONY: test scm git

test:
	 @export BZK_E2E_API_PORT=4000 ;\
     export BZK_E2E_SYSLOG_PORT=4001 ;\
     go test -v

scm: git

git:
	cd scm && docker build -t bazooka/e2e-git -f Dockerfile-git .
