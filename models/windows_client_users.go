package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ClientUser struct {
	Description string `json:"description"`
	Directory   string `json:"directory"`
	Gid         string `json:"gid"`
	GidSigned   string `json:"gid_signed"`
	Shell       string `json:"shell"`
	Type        string `json:"type"`
	UID         string `json:"uid"`
	UIDSigned   string `json:"uid_signed"`
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
	IsEnabled   bool   `json:"is_enabled"`
	LastLogon   string `json:"last_logon"`
}

type LocalUserGroup struct {
	Comment   string `json:"comment"`
	Gid       string `json:"gid"`
	GidSigned string `json:"gid_signed"`
	GroupSid  string `json:"group_sid"`
	Groupname string `json:"groupname"`
}

type NewLocalUser struct {
	Host                     []string `json:"host"`
	AccountDisabled          string   `json:"account_disabled"`
	Description              string   `json:"description"`
	Fullname                 string   `json:"fullname"`
	Group                    []string `json:"group"`
	HomeDirectory            string   `json:"home_directory"`
	LoginScript              string   `json:"login_script"`
	Username                 string   `json:"username"`
	Password                 string   `json:"password"`
	PasswordExpired          string   `json:"password_expired"`
	PasswordNeverExpires     string   `json:"password_never_expires"`
	Profile                  string   `json:"profile"`
	UserCannotChangePassword string   `json:"user_cannot_change_password"`
}

type appExecutionHistory struct {
	LastExecutionTime string `json:"last_execution_time"`
	Path              string `json:"path"`
}

func (sshConnection SshConnectionInfo) GetLocalUserLastLogonAndActive(userList []ClientUser) error {
	var err error
	for i := 0; i < len(userList); i++ {
		cmd := `net user ` + userList[i].Username + ` | findstr "Last active"`
		output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(cmd)
		if err != nil {
			return err
		}
		parseLocalUserLastLogonAndActiveResult(&userList[i], output)
		if strings.Trim(output, "\r\n ") == "" {
			return nil
		}
	}
	return err
}

func parseLocalUserLastLogonAndActiveResult(user *ClientUser, output string) {
	re := regexp.MustCompile(`\s{2,}`)
	lines := strings.Split(strings.Trim(output, "\r\n\t"), "\n")
	if strings.Contains(re.Split(lines[0], -1)[1], "Yes") {
		user.IsEnabled = true
	}
	user.LastLogon = re.Split(lines[1], -1)[1]
}

//Both Linux and Windows can use this
func (sshConnection *SshConnectionInfo) GetLocalUsers() ([]ClientUser, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM users WHERE type NOT LIKE 'roaming' AND type NOT LIKE 'special'"`)
	if err != nil {
		return nil, err
	}
	var userList []ClientUser

	err = json.Unmarshal([]byte(result), &userList)
	if err != nil {
		return nil, err
	}
	err = sshConnection.GetLocalUserLastLogonAndActive(userList)
	return userList, err
}

func AddNewWindowsUser(userJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/add_local_user.yml", userJson)

	return output, err
}

func DeleteWindowsUser(userJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/delete_local_user.yml", userJson)
	return output, err
}

func (sshConnection *SshConnectionInfo) GetWindowsGroupUserBelongTo(username string) ([]string, error) {
	isValid, err := regexp.MatchString("^[a-zA-Z0-9]+$", username)
	if !isValid || err != nil {
		return nil, err
	}
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT G.groupname FROM user_groups UG INNER JOIN users U ON U.uid=UG.uid INNER JOIN groups G ON G.gid=UG.GID WHERE U.username='` + username + `'`)
	if err != nil {
		return nil, err
	}
	type groupName struct {
		Groupname string `json:"groupname"`
	}
	var gNL []groupName
	var strGroupNameList []string
	err = json.Unmarshal([]byte(result), &gNL)
	for _, groupName := range gNL {
		strGroupNameList = append(strGroupNameList, groupName.Groupname)
	}
	return strGroupNameList, err
}

func (sshConnectionInfo *SshConnectionInfo) ReplaceWindowsGroupForUser(username string, group []string) (string, error) {
	type replacedGroup struct {
		Host     string   `json:"host"`
		Username string   `json:"username"`
		Group    []string `json:"group"`
	}

	groupListJson, err := json.Marshal(replacedGroup{Host: sshConnectionInfo.HostNameSSH, Username: username, Group: group})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/change_user_group_membership.yml", string(groupListJson))
	return output, err
}

func parseLoggedInUser(output string) ([]loggedInUser, error) {
	var loggedInUserList []loggedInUser

	var user loggedInUser
	re := regexp.MustCompile(`\s{2,}`)
	for i, line := range strings.Split(strings.Trim(output, "\n\r "), "\n") {
		if i == 0 {
			continue
		}
		line = strings.Trim(line, "\n\r ")
		vars := re.Split(line, -1)
		if len(vars) == 6 {
			user.Username = vars[0]
			user.SessionName = vars[1]
			user.SessionId = vars[2]
			user.State = vars[3]
			if vars[3] == "Disc" {
				user.State = "Disconnected"
			}
			user.IdleTime = vars[4]
			user.LogonTime = vars[5]
		} else if len(vars) == 5 {
			user.Username = vars[0]
			user.SessionId = vars[1]
			user.State = vars[2]
			if vars[2] == "Disc" {
				user.State = "Disconnected"
			}
			user.IdleTime = vars[3]
			user.LogonTime = vars[4]
		}
		loggedInUserList = append(loggedInUserList, user)
		user = loggedInUser{}
	}

	return loggedInUserList, nil
}

func (sshConnection SshConnectionInfo) GetLoggedInUsers() ([]loggedInUser, error) {
	var loggedInUserList []loggedInUser
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`quser`)
	if err != nil {
		if err.Error() == "Process exited with status 1" {
			return nil, nil
		}
		return loggedInUserList, err
	}
	loggedInUserList, err = parseLoggedInUser(result)
	return loggedInUserList, err
}

func (sshConnection SshConnectionInfo) KillWindowsLoginSession(sessionId int) error {
	_, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`logoff ` + strconv.Itoa(sessionId))
	if err != nil {
		return err
	}
	return err
}

func parseWindowsAppExecution(output string) ([]appExecutionHistory, error) {
	var appHistory []appExecutionHistory
	err := json.Unmarshal([]byte(output), &appHistory)
	return appHistory, err
}

func (sshConnection SshConnectionInfo) GetWindowsLoginAppExecutionHistory(username string) ([]appExecutionHistory, error) {
	var appHistory []appExecutionHistory
	query := fmt.Sprintf(`osqueryi --json "SELECT P.last_execution_time, P.path FROM background_activities_moderator AS P LEFT JOIN logged_in_users AS U WHERE P.sid=U.sid AND LOWER(U.user) = LOWER('%s') AND P.last_execution_time != ''"`, username)
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(query)
	if err != nil {
		return appHistory, err
	}
	appHistory, err = parseWindowsAppExecution(result)
	return appHistory, err
}

func (sshConnection SshConnectionInfo) checkIfWindowsUserEnabled(userList *[]ClientUser) (bool, error) {
	command := `powershell -Command "Get-LocalUser"`
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return false, err
	}
	parseUserEnabledResult(*userList, output)
	if strings.Trim(output, "\r\n ") == "" {
		return false, nil
	}
	if strings.Contains(output, "False") {
		return false, nil
	} else {
		return true, nil
	}
}

func parseUserEnabledResult(userList []ClientUser, output string) {
	output = strings.Trim(output, "\r\n ")
	for i, line := range strings.Split(output, "\r\n") {
		if i < 2 {
			continue
		}
		vars := strings.SplitN(line, " ", 2)
		for i, user := range userList {
			if strings.EqualFold(user.Username, strings.ToLower(strings.Trim(vars[0], "\r\n "))) {
				if strings.Contains(vars[1], "True") {
					userList[i].IsEnabled = true
					break
				} else {
					userList[i].IsEnabled = false
					break
				}
			}
		}
	}
}

func (sshConnection SshConnectionInfo) ChangeWindowsUserEnableStatus(username string, enabled bool) error {
	var isEnabled string
	if enabled {
		isEnabled = "yes"
	} else {
		isEnabled = "no"
	}
	command := fmt.Sprintf(`net user %s /active:%s"`, username, isEnabled)
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return err
	}
	if strings.Contains(output, "The command completed successfully.") {
		return nil
	}
	return errors.New(output)
}

func (sshConnection SshConnectionInfo) ChangeWindowsLocalUserPassword(username string, password string) error {
	type newPassword struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	nP, err := json.Marshal(newPassword{sshConnection.HostNameSSH, username, password})
	if err != nil {
		return err
	}
	_, err = RunAnsiblePlaybookWithjson("./yamls/windows_client/change_user_password.yml", string(nP))
	return err
}
