#!make
export version=v0.0.0-wip8

version:
	git tag ${version}
	git push origin ${version}

delete-tag:
	git push --delete origin ${version}
	git tag -d ${version}

deps:
	@echo "vendoring dependencies..."
	@go mod tidy
	@go mod vendor