package main

// Manifest represents a file holding additional information, which may be helpful
// or required to run user script (e.g., Python version, requirements etc.).
type Manifest struct{}

// FetchScript fetches named Script from ScriptRepo to ~/.ralph-cli dir.
func FetchScript(s Script) error {
	return nil
}

// EditScript opens an already fetched script for editing in editor given as
// $EDITOR env variable.
func EditScript(s Script) error {
	return nil
}

// CommitScript commits given Script (along with its Manifest) to ScriptRepo.
// Such script should be present in ~/.ralph-cli dir.
func CommitScript(s Script, m Manifest) (commitID string, e error) {
	return "dummyCommitID", nil
}

// CreateManifest creates an empty template for Manifest file associated with Script
// in ~/.ralph-cli dir.
func CreateManifest(s Script) error {
	return nil
}

// EditManifest opens for editing (in $EDITOR) a Manifest file associated with
// a given Script.
func EditManifest(s Script) error {
	return nil
}
