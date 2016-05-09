# This Makefile is meant only for releasing binaries hosted on GitHub
# For other cases, use standard Go tooling (i.e., go build, go install)
# and Glide (https://github.com/Masterminds/glide).

VERSION_FROM_GIT_TAG := `git describe --tags --abbrev=0 | sed 's/^v//'`

deps:
	glide install

release: clean deps
	rm -rf dist
	go get github.com/laher/goxc
	goxc -wc -pv=$(VERSION_FROM_GIT_TAG)
	@echo "Releasing binaries for supported platforms/OSs with version: $(VERSION_FROM_GIT_TAG)..."
	goxc -tasks-=go-install,go-vet,go-test
	@echo "Adding commit with updated '.goxc.json' file."
	git add .goxc.json
	git commit -m "Bumped PackageVersion in '.goxc.json'."
	@echo "Done."
	@echo "Remember to manually push release commits/tag to origin/master (with 'git push --follow-tags origin master')."

clean:
	rm -rf dist
