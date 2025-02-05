package models

import (
	"encoding/json"
	"strings"
)

type Package struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Source           string `json:"source"`
	Size             string `json:"size"`
	Arch             string `json:"arch"`
	PidWithNamespace int    `json:"pid_with_namespace"`
	MountNamespaceId string `json:"mount_namespace_id"`
}

type PackageJson struct {
	Host       []int    `json:"host"`
	HostString []string `json:"host_var"`
	File       string   `json:"file"`
	Mode       string   `json:"mode"`
	Package    string   `json:"package"`
	Link       string   `json:"link"`
	Page       int      `json:"page"`
}

type PackageReturn struct {
	Data         []Package `json:"data"`
	Current_page int       `json:"current_page"`
	Total_page   int       `json:"total_page"`
}

func ListAllPackge(hostList []int) ([]Package, error) {

	var (
		packageList []Package
		err         error
	)

	// Display installed package on one host
	if len(hostList) == 1 {
		packageList, err = GetInstalledPackagesOfOneHost(hostList[0])
		if err != nil {
			return packageList, err
		}

	} else if len(hostList) > 1 {
		// Display common package on many hosts
		for index, _ := range hostList {
			m := make(map[string]bool)
			var packageList1 []Package
			var packageList2 []Package
			if index == 0 {
				packageList1, err = GetInstalledPackagesOfOneHost(hostList[index])
				if err != nil {
					return packageList, err
				}
			} else {
				packageList1 = packageList
				packageList = []Package{}
			}
			packageList2, err = GetInstalledPackagesOfOneHost(hostList[index+1])
			if err != nil {
				return packageList, err
			}

			for _, packages := range packageList1 {
				m[packages.Name] = true
			}

			for _, packages := range packageList2 {
				if _, ok := m[packages.Name]; ok {
					packageList = append(packageList, Package{Name: packages.Name})
				}
			}
			if len(hostList) == index+2 {
				break
			}

		}
	}
	return packageList, err
}

func GetInstalledRPMPackages(sshConnection SshConnectionInfo) ([]Package, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM rpm_packages"`)
	if err != nil {
		return nil, err
	}
	var installedRPMs []Package

	err = json.Unmarshal([]byte(result), &installedRPMs)
	return installedRPMs, err
}

func GetInstalledDebPackages(sshConnection SshConnectionInfo) ([]Package, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM deb_packages"`)
	if err != nil {
		return nil, err
	}
	var installedDebs []Package

	err = json.Unmarshal([]byte(result), &installedDebs)
	return installedDebs, err
}

func GetInstalledPackagesOfOneHost(id int) ([]Package, error) {
	var packageList []Package
	sshConnection, err := GetSSHConnectionFromId(id)
	if err != nil {
		return packageList, err
	}

	// Verify os is debian or rpm
	if strings.Contains(sshConnection.OsType, "Ubuntu") {
		packageList, err = GetInstalledDebPackages(*sshConnection)
		if err != nil {
			return packageList, err
		}
	} else if strings.Contains(sshConnection.OsType, "CentOS") {
		packageList, err = GetInstalledRPMPackages(*sshConnection)
		if err != nil {
			return packageList, err
		}
	}
	return packageList, err
}

func PaginatePackageList(packageList []Package) map[int][]Package {
	offset := 0
	mapPaginate := make(map[int][]Package)
	index := 0
	for {
		data, currentoffset := Paginate(packageList, offset, len(packageList))
		mapPaginate[index] = data
		offset = currentoffset
		index += 1
		if currentoffset == len(packageList) {
			break
		}
	}
	return mapPaginate
}

func Paginate(x []Package, offset int, size int) ([]Package, int) {
	currentoffset := offset + 20
	if currentoffset > size {
		currentoffset = size
	}
	result := x[offset:currentoffset]

	return result, currentoffset
}

func ReturnPackgeList(mapPaginate map[int][]Package, page int) PackageReturn {
	var returnPackgeList PackageReturn
	returnPackgeList.Data = mapPaginate[page]
	returnPackgeList.Current_page = page
	returnPackgeList.Total_page = len(mapPaginate) - 1
	return returnPackgeList
}
