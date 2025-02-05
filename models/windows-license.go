package models

import (
	"encoding/hex"
	"strings"
)

type windowsLicense struct {
	ProductName string `json:"product_name"`
	ProductKey  string `json:"product_key"`
	ProductId   string `json:"product_id"`
}

func DecodeProductKey(digitalProductID []byte) string {
	var key string
	var keyOffset = 52
	var isWin8 = (digitalProductID[66] / 6) & 1
	digitalProductID[66] = (digitalProductID[66] & 0xf7) | (isWin8&2)*4
	var digits = "BCDFGHJKMPQRTVWXY2346789"
	var last = 0
	for i := 24; i >= 0; i-- {
		var current = 0
		for j := 14; j >= 0; j-- {
			current = current * 256
			current = int(digitalProductID[j+keyOffset]) + current
			digitalProductID[j+keyOffset] = byte(current / 24)
			current = current % 24
			last = current
		}
		key = string(digits[current]) + key
	}
	var keypart1 = key[1 : last+1]
	var keypart2 = key[last+1:]
	key = keypart1 + "N" + keypart2

	for i := 5; i < len(key); i += 6 {
		keypart1 = key[:i]
		keypart2 = key[i:]
		key = keypart1 + "-" + keypart2
	}
	return key
}

func (sshConnection SshConnectionInfo) GetWindowsLicenseKey() (windowsLicense, error) {
	var license windowsLicense

	var regKeyList []RegistryKey
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion' AND name = 'DigitalProductId'";`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	//regKeyList[0].Data = "F804000004000000380032003500300033002D00300033003300390031002D003000300030002D003000300030003000300030002D00300033002D0031003000330033002D00310039003000340032002E0030003000300030002D003300300039003200300032003100000000000000000000000000000000000000000000000000000000000000640034003500300035003900360066002D0038003900340064002D0034003900650030002D0039003600360061002D00660064003300390065006400340063003400630036003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000500072006F0050006C007500730056006F006C0075006D006500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003F0D00000000785DDBA04F09D2AD0800F5AB57A30FCE8D4745F1077031120F7995957058BD3CF03BBFE1442E5D3E5D8306858C96ACA44FFF18E4C98C3BB4C83E2A23EBCB15208FD980319AB3B757ECE25800320030002D0030003000350034003300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000056006F006C0075006D0065003A00470056004C004B00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006C0074004B004D00530000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	digitalProductID, err := hex.DecodeString(regKeyList[0].Data)
	if err != nil {
		return license, err
	}
	info, err := sshConnection.GetDetailSSHConInfo()
	if err != nil {
		return windowsLicense{}, nil
	}

	license.ProductName = info.OsName + " " + info.Architecture
	license.ProductKey = DecodeProductKey(digitalProductID)
	return license, err
}

func (sshConnection SshConnectionInfo) GetWindowsVmwareProductKey() (windowsLicense, error) {
	var regKeyList []RegistryKey
	var license windowsLicense
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT name,path FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\VMware, Inc.\VMware Workstation';`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	for _, key := range regKeyList {
		if strings.Contains(key.Name, "License.") {
			tmpString := key.Name
			result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT name,path,data FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\VMware, Inc.\VMware Workstation\` + key.Name + `';`)
			if err != nil {
				return license, err
			}
			regKeyList = nil
			regKeyList, err = parseKeyList(result)
			if err != nil {
				return license, err
			}
			if regKeyList == nil {
				return license, nil
			}
			for _, key := range regKeyList {
				if key.Name == "ProductID" {
					license.ProductName = key.Data + " " + tmpString
				}
				if key.Name == "Serial" {
					license.ProductKey = key.Data
				}
			}
			return license, err
		}
	}
	return license, err
}

func (sshConnection SshConnectionInfo) GetWindowsProductKey() (windowsLicense, error) {
	var regKeyList []RegistryKey
	var license windowsLicense
	tmpInfo, _ := sshConnection.GetDetailSSHConInfo()
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\SoftwareProtectionPlatform';"`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	for _, key := range regKeyList {
		if key.Name == "BackupProductKeyDefault" {
			license.ProductName = tmpInfo.OsName
			license.ProductKey = key.Data
			return license, err
		}
	}
	return license, err
}

func (sshConnection SshConnectionInfo) GetAllWindowsLicense() ([]windowsLicense, error) {
	var tmpLicense windowsLicense
	var licenseList []windowsLicense
	var err error
	tmpLicense, err = sshConnection.GetWindowsLicenseKey()
	if err != nil {
		return nil, err
	}
	licenseList = append(licenseList, tmpLicense)
	tmpLicense, err = sshConnection.GetWindowsVmwareProductKey()
	if err != nil {
		return nil, err
	}
	licenseList = append(licenseList, tmpLicense)
	return licenseList, err
}
