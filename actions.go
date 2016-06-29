package main

import (
	"fmt"
	"log"
	"net/http"
)

// PerformScan runs a scan of a given host using a script with scriptName.
// At this moment, we assume that only MAC addresses will be created/updated/deleted in Ralph.
func PerformScan(addrStr, scriptName string, dryRun bool, cfg *Config, cfgDir string) {
	if dryRun {
		// TODO(xor-xor): Wire up logger here.
		fmt.Println("INFO: Running in dry-run mode, no changes will be saved in Ralph.")
	}
	script, err := NewScript(scriptName, cfgDir)
	if err != nil {
		log.Fatalln(err)
	}
	if script.Manifest != nil && script.Manifest.Language == "python" && !VenvExists(script) {
		venvPath, err := CreatePythonVenv(script)
		if err != nil {
			log.Fatalln(err)
		}
		if err := InstallPythonReqs(venvPath, script); err != nil {
			log.Fatalln(err)
		}
	}
	addr, err := NewAddr(addrStr)
	if err != nil {
		log.Fatalln(err)
	}
	result, err := script.Run(addr, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	client, err := NewClient(cfg, addr, &http.Client{})
	if err != nil {
		log.Fatalln(err)
	}
	baseObj, err := addr.GetBaseObject(client)
	if err != nil {
		log.Fatalln(err)
	}
	oldEths, err := baseObj.GetEthernetComponents(client)
	// TODO(xor-xor): ExcludeMgmt should be removed when similar functionality will be implemented
	// in Ralph's API. Therefore, it should be considered as a temporary solution.
	oldEths, err = ExcludeMgmt(oldEths, addr, client)
	if err != nil {
		log.Fatalln(err)
	}
	var newEths []*EthernetComponent
	for _, mac := range result.MACAddresses {
		eth := NewEthernetComponent(mac, baseObj, "")
		newEths = append(newEths, eth)
	}
	diff, err := CompareEthernetComponents(oldEths, newEths)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		fmt.Println("No changes detected.")
		return
	}
	// When IP address is marked as "exposed in DHCP" in Ralph, then the only way to delete
	// EthernetComponent associated with its MAC addres is through a suitable transition
	// from Ralph's GUI (i.e., it is not possible via REST API by desing). Therefore, we need
	// to exclude such EthernetComponents from diff.Delete.
	if len(diff.Delete) > 0 {
		diff, err = ExcludeExposedInDHCP(diff, client, false)
		if err != nil {
			log.Fatalln(err)
		}
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		log.Fatalln(err)
	}
}