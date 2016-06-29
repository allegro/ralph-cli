# This Makefile is meant only for releasing binaries hosted on GitHub.
# For other cases, use standard Go tooling (i.e., go build, go install)
# and Glide (https://github.com/Masterminds/glide).
#
# Usage (assuming that we do not use any release branches):
# git checkout master
# git merge develop
# make release VERSION=0.1.0

deps:
	glide install

release: clean deps
	go get github.com/laher/goxc
	goxc -wc -pv=$$VERSION
	@echo "Adding commit with updated '.goxc.json' file..."
	git add .goxc.json
	git commit -m "Bumped PackageVersion in .goxc.json to $$VERSION."
	@echo "Adding release tag..."
	git tag -a -m "Release of ralph-cli v$$VERSION." v$$VERSION
	@echo "Pushing changes to origin..."
	git push --follow-tags origin master
	@echo "Releasing binaries for supported platforms/OSs with version: $$VERSION..."
	goxc -tasks-=go-install,go-vet,go-test
	@echo "Done."

clean:
	rm -rf dist
