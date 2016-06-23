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
	@echo "Adding commit with updated '.goxc.json' file."
	git add .goxc.json
	git commit -m "Bumped PackageVersion in .goxc.json to $$VERSION."
	git tag -a -m "Release of ralph-cli $$VERSION." $$VERSION
	@echo "Releasing binaries for supported platforms/OSs with version: $$VERSION..."
	goxc -tasks-=go-install,go-vet,go-test
	@echo "Done."
	@echo "Remember to manually push release commits/tag to origin/master (with 'git push --follow-tags origin master')."

clean:
	rm -rf dist
