package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

var macs = PopulateMACs()

// PopulateMACs create some fake MAC addresses for use with tests. This is just a
// convenience function, which allows writing something like macs["aa:aa:aa:aa:aa:aa"]
// instead of more elaborate MACaddress literals etc.
func PopulateMACs() map[string]MACAddress {
	macs := make(map[string]MACAddress)
	ss := []string{
		"aa:aa:aa:aa:aa:aa",
		"aa:bb:cc:dd:ee:ff",
		"a1:b2:c3:d4:e5:f6",
		"74:86:7a:ee:20:e8",
	}
	for _, s := range ss {
		hwAddr, _ := net.ParseMAC(s)
		macs[s] = MACAddress{hwAddr}
	}
	return macs
}

// GetTempCfgDir creates a temporary directory structure /SYSTEM/TEMP/DIR/ralph-cli-tests-RANDOM/.ralph-cli,
// returns its full path as cfgDir, and also a path to /SYSTEM/TEMP/DIR/ralph-cli-tests-RANDOM as baseDir.
// It is the caller's reposnsibility to remove this whole structure when no longer needed (although it will
// remove itself in case of GetTempCfgDir's failure).
func GetTempCfgDir() (cfgDir, baseDir string, err error) {
	baseDir, err = ioutil.TempDir("", "ralph-cli-tests-")
	if err != nil {
		goto FAIL
	}
	cfgDir, err = GetCfgDirLocation(baseDir)
	if err != nil {
		goto FAIL
	}
	if err != nil {
		goto FAIL
	}
	err = PrepareCfgDir(cfgDir, "config.toml")
	if err != nil {
		goto FAIL
	}
	return cfgDir, baseDir, nil

FAIL:
	os.RemoveAll(baseDir)
	return "", "", err
}

// GetHelperCommand returns a function for mocking exec.Command. This function creates a helper process
// given as name arg, and only this process should contain the actual mock's logic. This idea is taken
// from Golang's stdlib, see: https://github.com/golang/go/blob/master/src/os/exec/exec_test.go#L32.
func GetHelperCommand(name string) func(command string, args ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		helperProcess := fmt.Sprintf("-test.run=%s", name)
		cs := []string{helperProcess, "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

// LoadFixture reads a fixture file given by dir & file and returns it as a string.
func LoadFixture(dir, file string) (string, error) {
	path, err := filepath.Abs(filepath.Join(dir, file))
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MockServerClient creates fake HTTP server and client, with code and body that should
// be returned by server.
func MockServerClient(code int, body string) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	client := &Client{
		scannedAddr: Addr(""),
		ralphURL:    server.URL,
		apiKey:      "",
		client:      httpClient,
	}

	return server, client
}

// TestEqByte is a predicate function testing two byte slices for equality.
func TestEqByte(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestEqStr is a predicate function testing two string slices for equality.
func TestEqStr(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FakeComponent is meant only for tests.
type FakeComponent struct {
	FakeField string
}

// String implements Component interface.
func (f FakeComponent) String() string {
	return "fake component"
}

// IsEqualTo implements Component interface.
func (f FakeComponent) IsEqualTo(c Component) bool {
	return false
}

// PtrToStr returns a pointer to a string s, which may be useful for
// constructing struct literals with *string fields.
func PtrToStr(s string) *string {
	return &s
}

// PtrToInt returns a pointer to an int i, which may be useful for constructing
// struct literals with *int fields.
func PtrToInt(i int) *int {
	return &i
}
