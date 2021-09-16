package models

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
	"golang.org/x/crypto/ssh"
)

type SshConnectionInfo struct {
	SSHConnectionId int    `json:"sshConnectionId"`
	UserSSH         string `json:"userSSH"`
	PasswordSSH     string `json:"passwordSSH"`
	HostNameSSH     string `json:"hostnameSSH"`
	HostSSH         string `json:"hostSSH"`
	PortSSH         int    `json:"portSSH"`
	CreatorId       int    `json:"creatorId"`
	SSHKeyId        int    `json:"sshKeyId"`
	OsType          string `json:"osType"`
}

//Read private key from private key file
func ProcessPrivateKey(keyId int) (ssh.AuthMethod, error) {
	//buffer, err := ioutil.ReadFile(file)
	//fmt.Println(string(buffer))
	privateKey, _ := GetSSHKeyFromId(keyId)
	decrytedPrivateKey, err := AESDecryptKey(privateKey.PrivateKey)
	if err != nil {
		return nil, err
	}

	buffer := []byte(decrytedPrivateKey)
	// if err != nil {
	// 	return nil, err
	// }

	key, err := ssh.ParsePrivateKey(buffer)
	//If error means Private key is protected by passphrase
	if err != nil {
		utils.EnvInit()
		key, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(os.Getenv("SECRET_SSH_PASSPHRASE")))
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeys(key), err
	}
	return ssh.PublicKeys(key), err
}

//Test the SSH connection using private key
func (sshConnection *SshConnectionInfo) TestConnectionPublicKey() (bool, error) {
	//If private key is incorrect or wrong format, return error immediately
	var auth []ssh.AuthMethod
	authMethod, err := ProcessPrivateKey(sshConnection.SSHKeyId)
	if err != nil {
		return false, err
	}
	//Else continue testing connection using the above key
	auth = append(auth, authMethod)

	sshConfig := &ssh.ClientConfig{
		User:            sshConnection.UserSSH,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	_, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return false, err
	} else {
		return true, err
	}
}

func (sshConnection *SshConnectionInfo) AddSSHConnectionToDB() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	encryptedPassword := AESEncryptKey(sshConnection.PasswordSSH)

	stmt, err := db.Prepare("INSERT INTO ssh_connections (sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id) VALUES (?,?,?,?,?,?,?)")
	if err != nil {

		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sshConnection.UserSSH, encryptedPassword, sshConnection.HostSSH, sshConnection.HostNameSSH, sshConnection.PortSSH, sshConnection.CreatorId, sshConnection.SSHKeyId)
	if err != nil {

		return false, err
	}

	return true, err
}

func GetAllSSHConnection() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &connectionInfo.SSHKeyId)
		if err != nil {
			return nil, err
		}

		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllSSHConnectionWithPassword() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_host, sc_hostname, sc_password, sc_port, creator_id, ssh_key_id, sc_ostype 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PasswordSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &connectionInfo.SSHKeyId, &connectionInfo.OsType)
		if err != nil {
			return nil, err
		}

		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetSSHConnectionFromHostName(sshHostName string) (*SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	//var encryptedPassword string
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_host, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_hostname = ?", sshHostName)
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &sshConnection.HostSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &sshConnection.SSHKeyId)
	if row == nil {
		return nil, errors.New("ssh connection doesn't exist")
	}

	/*sshConnection.PasswordSSH = AESDecryptKey(encryptedPassword)
	if err != nil {
		return nil, errors.New("fail to retrieve ssh connection info")
	}*/
	return &sshConnection, err
}

func GetSSHConnectionFromId(sshConnectionId int) (*SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	var encryptedPassword string
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_connection_id = ?", sshConnectionId)
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &encryptedPassword, &sshConnection.HostSSH, &sshConnection.HostNameSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &sshConnection.SSHKeyId)
	if row == nil {
		return nil, errors.New("ssh connection doesn't exist")
	}
	if err != nil {
		return nil, errors.New("fail to retrieve ssh connection info")
	}

	sshConnection.PasswordSSH, err = AESDecryptKey(encryptedPassword)
	if err != nil {
		return nil, errors.New("fail to decrypt ssh connection password")
	}
	return &sshConnection, err
}

func GetSshHostnameFromId(sshConnectionId int) (string, error) {
	db := database.ConnectDB()
	defer db.Close()

	var hostname string
	row := db.QueryRow("SELECT sc_hostname FROM ssh_connections WHERE sc_connection_id = ?", sshConnectionId)
	err := row.Scan(&hostname)
	if row == nil {
		return hostname, errors.New("ssh connection doesn't exist")
	}

	return hostname, err
}

func GetSSHConnectionFromIP(ip string) (SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	var encryptedPassword string
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_host = ?", ip)
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &encryptedPassword, &sshConnection.HostSSH, &sshConnection.HostNameSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &sshConnection.SSHKeyId)
	if row == nil {
		return sshConnection, errors.New("ssh connection doesn't exist")
	}
	if err != nil {
		return sshConnection, errors.New("fail to retrieve ssh connection info")
	}
	return sshConnection, err
}

// Check Public Key of user exist or not
func (sshConnection *SshConnectionInfo) IsKeyExist() bool {
	if _, err := GetSSHKeyFromId(sshConnection.SSHKeyId); err == nil {
		return true
	} else {
		return false
	}
}

//Delete SSH Connection Function
func DeleteSSHConnection(id int) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM ssh_connections WHERE sc_connection_id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, errors.New("no SSH Connections with this ID exists")
	}
	return true, err
}

//Get SSH Connection and run command on it
func RunCommandFromSSHConnection(sshConnection SshConnectionInfo, command string) (string, error) {
	result, err := ExecCommand(command, sshConnection.UserSSH, sshConnection.PasswordSSH, sshConnection.HostSSH, sshConnection.PortSSH)
	return result, err
}

func connectSSH(user, password, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	// get auth method

	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect to ssh

	addr = fmt.Sprintf("%s:%d", host, port)

	sshClient, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return sshClient, err
	}

	return sshClient, nil
}

func ExecCommand(cmd string, userSSH string, passwordSSH string, hostSSH string, portSSH int) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	//create ssh connect
	sshClient, err = connectSSH(userSSH, passwordSSH, hostSSH, portSSH)
	if err != nil {
		return "Wrong username or password to connect remote server", err
	} else {
		defer sshClient.Close()
		//create a session. It is one session per command
		session, err = sshClient.NewSession()
		if err != nil {
			return "Failed to open new session", err
		}
		defer session.Close()
		var b bytes.Buffer //import "bytes"
		session.Stdout = &b
		err = session.Run(cmd)
		return b.String(), err
	}

}

func GenerateInventory() error {
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return err
	}
	var inventory string
	for _, sshConnection := range sshConnectionList {
		line := sshConnection.HostNameSSH + " ansible_host=" + sshConnection.HostSSH + " ansible_port=" + fmt.Sprint(sshConnection.PortSSH) + " ansible_user=" + sshConnection.UserSSH + "\n"
		inventory += line
	}

	err = ioutil.WriteFile("/etc/ansible/hosts", []byte(inventory), 0644)
	//fmt.Println(err.Error())
	return err
}

//Run command through SSH using SSH keys
func (sshConnection *SshConnectionInfo) RunCommandFromSSHConnectionUseKeys(command string) (string, error) {
	result, err := sshConnection.ExecCommandWithSSHKey(command)
	return result, err
}

func (sshConnection *SshConnectionInfo) connectSSHWithSSHKeys() (*ssh.Client, error) {
	//If private key is incorrect or wrong format, return error immediately
	var auth []ssh.AuthMethod
	authMethod, err := ProcessPrivateKey(sshConnection.SSHKeyId)
	if err != nil {
		return nil, err
	}
	//Else continue testing connection using the above key
	auth = append(auth, authMethod)

	sshConfig := &ssh.ClientConfig{
		User:            sshConnection.UserSSH,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return sshClient, err
	}

	return sshClient, nil
}

func (sshConnection *SshConnectionInfo) ExecCommandWithSSHKey(cmd string) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	//create ssh connect
	sshClient, err = sshConnection.connectSSHWithSSHKeys()
	if err != nil {
		return "Wrong username or password to connect remote server", err
	} else {
		defer sshClient.Close()
		//create a session. It is one session per command
		session, err = sshClient.NewSession()
		if err != nil {
			return "Failed to open new session", err
		}
		defer session.Close()
		var b bytes.Buffer //import "bytes"
		session.Stdout = &b
		err = session.Run(cmd)
		return b.String(), err
	}
}

// Get OS Type of PC
func (sshConnection *SshConnectionInfo) GetOsType() (string, error) {
	var (
		osType string
		err    error
	)

	// Initialize extra value and run yaml file
	var extraValue map[string]string = map[string]string{"host": sshConnection.HostNameSSH}
	output, err := LoadYAML("./yamls/checkOsType.yml", extraValue)
	if err != nil {
		return osType, errors.New("fail to load yaml file")
	}

	// Retrieving value from Json format
	value := ExtractJsonValue(output, []string{"msg"})
	osType = value[0]

	// Convert friendly name for windows type
	if strings.Contains(osType, "Windows") {
		osType = "Windows"
	}
	return osType, err
}

// Update Os Type to DB
func (sshconnection *SshConnectionInfo) UpdateOsType() error {
	db := database.ConnectDB()
	defer db.Close()

	query := "UPDATE ssh_connections SET sc_ostype = ? WHERE sc_hostname = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sshconnection.OsType, sshconnection.HostNameSSH)
	if err != nil {
		return err
	}
	return err
}

func (sshConnection *SshConnectionInfo) GetIptables() ([]IptableRule, error) {
	var iptables []IptableRule
	firewallRule, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM iptables"`)
	if err != nil {
		return iptables, err
	}
	iptables, err = ParseIptables(firewallRule)
	return iptables, err
}

func (sshConnection *SshConnectionInfo) GetWindowsFirewall(direction string) ([]PortNetshFirewallRule, error) {
	var firewallRules []PortNetshFirewallRule
	firewallRule, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`netsh advfirewall firewall show rule name=all dir="` + direction)
	if err != nil {
		return firewallRules, err
	}
	firewallRules, err = ParsePortNetshFirewallRuleFromPowershell(firewallRule)
	return firewallRules, err
}
