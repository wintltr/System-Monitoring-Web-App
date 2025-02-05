package models

import (
	"strings"

	"github.com/wintltr/login-api/database"
)

type SysInfo struct {
	ConnectionId int    `json:"id"`
	HostnameSSH  string `json:"hostnameSSH"`
	AvgCPU       string `json:"avgcpu"`
	AvgMem       string `json:"avgmem"`
	Timestamp    string `json:"timestamp"`
	State        string `json:"state"`
	OsType       string `json:"osType"`
}

type OnlineStatus struct {
	ConnectionId int
	IsOn         bool
}

func InsertSysInfoToDB(sysInfo SysInfo, ip string, hostname string, connectionId int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO sys_info_logs (syl_hostname, syl_avg_cpu, syl_avg_mem, syl_timestamp, syl_connection_id) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(hostname, sysInfo.AvgCPU, sysInfo.AvgMem, sysInfo.Timestamp, connectionId)
	if err != nil {
		return err
	}
	return err
}

func GetLatestSysInfo(sshConnection SshConnectionInfo) (SysInfo, error) {
	var (
		sysInfo SysInfo
		err     error
	)
	sysInfo.ConnectionId = sshConnection.SSHConnectionId
	sysInfo.HostnameSSH = sshConnection.HostNameSSH

	// Linux CPU and Memory
	if strings.Contains(sshConnection.OsType, "CentOS") || strings.Contains(sshConnection.OsType, "Ubuntu") {
		sysInfo.AvgMem = CalcAvgMemUseForLinux(sshConnection)
		sysInfo.AvgCPU, err = CalcAvgCPUFromTopForLinux(sshConnection)
		if err != nil {
			err = nil
			sysInfo.AvgCPU = ""
		}
	} else if strings.Contains(sshConnection.OsType, "Windows") {
		sysInfo.AvgCPU, err = CalcAvgCPUFromTopForWindows(sshConnection)
		if err != nil {
			err = nil
			sysInfo.AvgCPU = ""
		}
	}

	return sysInfo, err
}

func CalcAvgMemUseForLinux(sshConnection SshConnectionInfo) string {
	var (
		result string
		err    error
	)
	command := "free | grep Mem | awk '{print $3/$2 * 100.0}'"
	result, err = sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return result
	}
	return strings.Trim(string(result), "\n")
}

func CalcAvgCPUFromTopForLinux(sshConnection SshConnectionInfo) (string, error) {
	var (
		//cpuUse float32
		result string
		err    error
	)

	//command := "top -b -n 1"
	command := `grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage " % "}'`
	result, err = sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	/*
			lines := strings.Split(string(result), "\n")
			for _, line := range lines {
				if strings.Contains(line, "%Cpu(s):") {
					atributes := strings.Split(line, ",")
					idle, err := strconv.ParseFloat(strings.Trim((atributes[3][:5]), " "), 32)
					if err != nil {
						return "", err
					}
					cpuUse = 100 - float32(idle)
				}
			}
		return fmt.Sprintf("%.1f", cpuUse), err
	*/
	return result, err
}

func CalcAvgCPUFromTopForWindows(sshConnection SshConnectionInfo) (string, error) {
	var (
		//cpuUse float32
		result string
		err    error
	)

	command := `@for /f "skip=1" %p in ('wmic cpu get loadpercentage') do @echo %p%`
	result, err = sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(strings.Split(result, "\n")[0])

	return result, err
}

func GetAllSysInfo(sshConnectionList []SshConnectionInfo) ([]SysInfo, error) {
	var sysInfoList []SysInfo
	var sysInfo SysInfo
	var err error

	for _, sshConnection := range sshConnectionList {
		sysInfo, err = GetLatestSysInfo(sshConnection)
		if err != nil {
			return sysInfoList, err
		}

		// Append OsType and State of machine to sysInfo
		sysInfo.OsType = sshConnection.OsType
		sysInfo.State, err = GetPcStateByID(sshConnection.SSHConnectionId)
		if err != nil {
			return sysInfoList, err
		}
		sysInfoList = append(sysInfoList, sysInfo)
	}

	return sysInfoList, nil
}

func CheckOnlineStatus(sshConnectionlist []SshConnectionInfo) []OnlineStatus {
	var statuses []OnlineStatus
	for _, sshConnection := range sshConnectionlist {
		sysinfo, _ := GetLatestSysInfo(sshConnection)
		if sysinfo.AvgCPU == "" {
			statuses = append(statuses, OnlineStatus{sshConnection.SSHConnectionId, false})
		} else {
			statuses = append(statuses, OnlineStatus{sshConnection.SSHConnectionId, true})
		}
	}
	return statuses
}
