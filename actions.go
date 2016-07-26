package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// PerformScan runs a scan of a given host using a script with scriptName.
func PerformScan(addrStr, scriptName string, components map[string]bool, withBIOSAndFirmware, withModel, dryRun bool, cfg *Config, cfgDir string) {
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

	var changesDetected bool
	if components["eth"] || components["all"] {
		if changed := getEthernets(addr, result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if components["mem"] || components["all"] {
		if changed := getMemory(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if components["fcc"] || components["all"] {
		if changed := getFibreChannelCards(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if components["cpu"] || components["all"] {
		if changed := getProcessors(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if components["disk"] || components["all"] {
		if changed := getDisks(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if withBIOSAndFirmware {
		if changed := getBIOSAndFirmwareVersions(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}
	if withModel {
		if changed := getModelName(result, baseObj, client, dryRun); changed {
			changesDetected = true
		}
	}

	if !changesDetected {
		fmt.Println("No changes detected.")
	}
}

func getEthernets(addr Addr, result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	oldEths, err := baseObj.GetEthernets(client)
	// TODO(xor-xor): ExcludeMgmt should be removed when similar functionality
	// will be implemented in Ralph's API. Therefore, it should be considered as
	// a temporary solution.
	oldEths, err = ExcludeMgmt(oldEths, addr, client)
	if err != nil {
		log.Fatalln(err)
	}
	var newEths []*Ethernet
	for i := 0; i < len(result.Ethernets); i++ {
		result.Ethernets[i].BaseObject = *baseObj
		newEths = append(newEths, &result.Ethernets[i])
	}
	diff, err := CompareEthernets(oldEths, newEths)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		return false
	}
	// When IP address is marked as "exposed in DHCP" in Ralph, then the only
	// way to delete Ethernet associated with its MAC address is through a so
	// called "transition". Therefore, we need to exclude such Ethernets from
	// diff.Delete.
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
	return true
}

func getMemory(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	oldMem, err := baseObj.GetMemory(client)
	if err != nil {
		log.Fatalln(err)
	}
	var newMem []*Memory
	for i := 0; i < len(result.Memory); i++ {
		result.Memory[i].BaseObject = *baseObj
		newMem = append(newMem, &result.Memory[i])
	}

	diff, err := CompareMemory(oldMem, newMem)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		return false
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		log.Fatalln(err)
	}
	return true
}

// ExcludeMgmt filters eths by excluding Ethernets associated with given IP
// address, but only when such address is a management one.
// This function should be considered as a temporary solution, and will be removed once
// similar functionality will be implemented in Ralph's API.
func ExcludeMgmt(eths []*Ethernet, ip Addr, c *Client) ([]*Ethernet, error) {
	var ethsFiltered []*Ethernet
	addrs, err := getIPAddresses(fmt.Sprintf("address=%s", ip), c)
	if err != nil {
		return nil, err
	}
	// IP addresses are unique in Ralph, so there's no need to check for addrs.Count > 1.
	if addrs.Count == 0 || !addrs.Results[0].IsMgmt {
		return eths, nil
	}
	for _, eth := range eths {
		if eth.ID != addrs.Results[0].Ethernet.ID {
			ethsFiltered = append(ethsFiltered, eth)
		}
	}
	return ethsFiltered, nil
}

// ExcludeExposedInDHCP takes Diff, and examines Ethernets from d.Delete
// list. In quite unlikely, but possible case of finding such Ethernet, it will
// excluded from said diff, and warning message will be printed for user (if
// noOutput is set to true, then no message will be printed - this is meant for
// testing).
func ExcludeExposedInDHCP(diff *Diff, c *Client, noOutput bool) (*Diff, error) {
	var ethsFiltered []*DiffComponent
	for _, d := range diff.Delete {
		switch ec := d.Component.(type) {
		case *Ethernet:
			ip, err := checkIfExposedInDHCP(&ec.MACAddress, c)
			if err != nil {
				return nil, err
			}
			if ip.Address != "" {
				if !noOutput {
					fmt.Printf("WARNING: Ethernet with MAC address %s cannot be deleted, "+
						"because IP address associated with it (%s) is marked as \"exposed in DHCP\" "+
						"in Ralph. Please use a suitable transition from Ralph's GUI for that.\n",
						ec.MACAddress.String(), ip.Address) // TODO(xor-xor): Use logger instead.
				}
				continue
			}
		default:
			return nil, errors.New("unknown type in Ethernet context (ExcludeExposedInDHCP function)")
		}
		ethsFiltered = append(ethsFiltered, d)
	}
	diff.Delete = ethsFiltered
	return diff, nil
}

// checkIfExposedInDHCP is a helper function for ExcludeExposedInDHCP.
func checkIfExposedInDHCP(m *MACAddress, c *Client) (IPAddress, error) {
	addrs, err := getIPAddresses(fmt.Sprintf("ethernet__mac=%s", m.String()), c)
	if err != nil {
		return IPAddress{}, err
	}
	for _, ip := range addrs.Results {
		if ip.ExposeInDHCP == true {
			return ip, nil
		}
	}
	return IPAddress{}, nil
}

func getFibreChannelCards(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	oldFCC, err := baseObj.GetFibreChannelCards(client)
	if err != nil {
		log.Fatalln(err)
	}

	var newFCC []*FibreChannelCard
	for i := 0; i < len(result.FibreChannelCards); i++ {
		result.FibreChannelCards[i].BaseObject = *baseObj
		newFCC = append(newFCC, &result.FibreChannelCards[i])
	}

	diff, err := CompareFibreChannelCards(oldFCC, newFCC)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		return false
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		log.Fatalln(err)
	}
	return true
}

func getProcessors(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	oldProcs, err := baseObj.GetProcessors(client)
	if err != nil {
		log.Fatalln(err)
	}

	var newProcs []*Processor
	for i := 0; i < len(result.Processors); i++ {
		result.Processors[i].BaseObject = *baseObj
		newProcs = append(newProcs, &result.Processors[i])
	}

	diff, err := CompareProcessors(oldProcs, newProcs)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		return false
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		log.Fatalln(err)
	}
	return true
}

func getDisks(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	oldDisks, err := baseObj.GetDisks(client)
	if err != nil {
		log.Fatalln(err)
	}

	var newDisks []*Disk
	for i := 0; i < len(result.Disks); i++ {
		result.Disks[i].BaseObject = *baseObj
		newDisks = append(newDisks, &result.Disks[i])
	}

	diff, err := CompareDisks(oldDisks, newDisks)
	if err != nil {
		log.Fatalln(err)
	}
	if diff.IsEmpty() {
		return false
	}
	_, err = SendDiffToRalph(client, diff, dryRun, false)
	if err != nil {
		log.Fatalln(err)
	}
	return true
}

// Hence DataCenterAsset has only two fields on ralph-cli's side and we send it
// to Ralph only as an update (i.e., with PATCH method), there's no need for a
// separate, more sophisticated function like CompareDataCenterAssets.
func getBIOSAndFirmwareVersions(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	dcAsset, err := baseObj.GetDataCenterAsset(client)
	if err != nil {
		log.Fatalln(err)
	}
	// Setting nil to any DataCenterAsset field effectively excludes it from
	// JSON sent to Ralph.
	dcAsset.Remarks = nil
	var changed bool
	if result.FirmwareVersion != *dcAsset.FirmwareVersion {
		*dcAsset.FirmwareVersion = result.FirmwareVersion
		changed = true
	} else {
		dcAsset.FirmwareVersion = nil
	}
	if result.BIOSVersion != *dcAsset.BIOSVersion {
		*dcAsset.BIOSVersion = result.BIOSVersion
		changed = true
	} else {
		dcAsset.BIOSVersion = nil
	}
	if changed {
		var diff Diff
		d, err := NewDiffComponent(dcAsset)
		if err != nil {
			log.Fatalln(err)
		}
		diff.Update = append(diff.Update, d)
		_, err = SendDiffToRalph(client, &diff, dryRun, false)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return changed
}

func getModelName(result *ScanResult, baseObj *BaseObject, client *Client, dryRun bool) bool {
	if result.ModelName == "" {
		return false
	}

	dcAsset, err := baseObj.GetDataCenterAsset(client)
	if err != nil {
		log.Fatalln(err)
	}
	// exclude these fields from JSON sent to Ralph
	dcAsset.FirmwareVersion = nil
	dcAsset.BIOSVersion = nil

	const remarkTemplate = ">>> ralph-cli: detected model name: %s <<<"
	r, err := regexp.Compile(">>> ralph-cli: detected model name:.*<<<")
	if err != nil {
		log.Fatalln(err)
	}
	newRemark := fmt.Sprintf(remarkTemplate, result.ModelName)
	var changed bool
	switch oldRemark := r.FindString(*dcAsset.Remarks); {
	case oldRemark == newRemark:
		return false
	case oldRemark != "": // replace existing remark
		*dcAsset.Remarks = r.ReplaceAllString(*dcAsset.Remarks, newRemark)
		changed = true
	default: // no existing remark, append one
		var separator string
		if len(*dcAsset.Remarks) > 0 {
			separator = "\n"
		}
		*dcAsset.Remarks = strings.Join([]string{
			*dcAsset.Remarks,
			fmt.Sprintf(remarkTemplate, result.ModelName),
		}, separator)
		changed = true
	}

	if changed {
		var diff Diff
		d, err := NewDiffComponent(dcAsset)
		if err != nil {
			log.Fatalln(err)
		}
		diff.Update = append(diff.Update, d)
		_, err = SendDiffToRalph(client, &diff, dryRun, false)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return changed
}
