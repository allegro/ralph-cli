package main

import (
	"fmt"
	"net/http"
	"os"
)

// PerformScan runs a scan of a given host using a set of scripts.
// At this moment, we assume that there will be only one script here (idrac.py),
// and that only MAC addresses will be created/updated/deleted in Ralph.
func PerformScan(addrStr string, scripts []string, dryRun bool, cfgDir string) {
	if dryRun {
		fmt.Println("Running in dry-run mode, no changes will be saved in Ralph.")
	}
	script, err := NewScript(scripts[0], cfgDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	addr, err := NewAddr(addrStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	result, err := script.Run(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	client, err := NewClient(
		os.Getenv("RALPH_API_URL"),
		os.Getenv("RALPH_API_KEY"),
		addr,
		&http.Client{},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	baseObj, err := addr.GetBaseObject(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	oldEths, err := baseObj.GetEthernetComponents(client)
	// TODO(xor-xor): ExcludeMgmt should be removed when similar functionality will be implemented
	// in Ralph's API. Therefore, it should be considered as a temporary solution.
	oldEths, err = ExcludeMgmt(oldEths, addr, client)
	if err != nil {
		fmt.Println(err)
		return
	}
	var newEths []*EthernetComponent
	for _, mac := range result.MACAddresses {
		eth := NewEthernetComponent(mac, baseObj, "")
		newEths = append(newEths, eth)
	}
	diff, err := CompareEthernetComponents(oldEths, newEths)
	if err != nil {
		fmt.Println(err)
	}
	if diff.IsEmpty() {
		fmt.Println("No changes detected.")
		return
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		fmt.Println(err)
	}
}
