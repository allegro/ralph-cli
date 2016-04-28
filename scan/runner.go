package scan

import "fmt"

type ScanResult struct {
	macAddresses     []string `json:"mac_addresses"`
	hostname         string   `json:"hostname"`
	systemCoresCount int      `json:"system_cores_count"`
	systemMemory     int      `json:"system_memory"`
	serialNumber     string   `json:"serial_number"`
	systemIpAddress  []string `json:"system_ip_addresses"`
}

func ScanRunner(ips *[]string, components *[]string, plugins *[]string, dryRun *bool, deviceTemplate *string) {
	if len(*ips) == 0 {
		fmt.Println("IP address not set, trying scan 127.0.0.1")
		*ips = []string{"127.0.0.1"}
	}

	// EXAMPLE SCAN RESULT

	macAddresses := []string{"00:0A:E6:3E:FD:E1"}
	ipAdresses := []string{"10.72.21.1", "10.133.21.65"}

	for _, ip := range *ips {
		result := ScanResult{
			macAddresses:     macAddresses,
			hostname:         "ralph.local",
			systemCoresCount: 4,
			systemMemory:     4096,
			serialNumber:     "SAS7F8S7A557A7",
			systemIpAddress:  ipAdresses,
		}

		fmt.Printf("\nStarted scan for %s\n", ip)
		fmt.Printf("\nExample scan result for %s\n%+v\n", ip, result)
	}
}
