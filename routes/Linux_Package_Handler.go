package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func PackageRemove(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	// Retrieve Json Format
	var packages models.PackageJson

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}
	json.Unmarshal(reqBody, &packages)

	// Call function Load in yaml.go
	var host []string
	for _, id := range packages.Host {
		hostname, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse id").Error())
			return
		}
		host = append(host, hostname.HostNameSSH)
	}
	packages.HostString = host

	// Marshal
	packetJsonMarshal, err := json.Marshal(packages)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse marshal json").Error())
		return
	}

	var eventStatus string
	output, err := models.RunAnsiblePlaybookWithjson("./yamls/"+packages.File, string(packetJsonMarshal))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to load yaml")
		eventStatus = "failed"
	} else {
		// Processing Output From Ansible
		status, fatalList, err := models.ProcessingAnsibleOutput(output)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		var statusCode int
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatalList)
		if len(fatalList) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}
	// Write Event Web
	description := "Package \"" + packages.Package + "\" removed from " + fmt.Sprint(host) + " " + eventStatus
	_, err = models.WriteWebEvent(r, "Package", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}
func PackageInstall(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	// Retrieve Json Format
	var packages models.PackageJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}
	json.Unmarshal(reqBody, &packages)

	// Load File Yaml Install
	var host []string
	for _, id := range packages.Host {
		hostname, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse id").Error())
			return
		}
		host = append(host, hostname.HostNameSSH)
	}
	packages.HostString = host

	// Marshal
	packetJsonMarshal, err := json.Marshal(packages)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse marshal json").Error())
		return
	}
	var eventStatus string
	output, err := models.RunAnsiblePlaybookWithjson("./yamls/"+packages.File, string(packetJsonMarshal))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to load yaml")
		eventStatus = "failed"
	} else {

		// Processing Output From Ansible
		status, fatalList, err := models.ProcessingAnsibleOutput(output)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		var statusCode int
		returnJson := simplejson.New()
		returnJson.Set("Fatal", fatalList)
		returnJson.Set("Status", status)
		if len(fatalList) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}
	// Write Event Web
	description := "Package \"" + packages.Package + "\" installed to" + fmt.Sprint(host) + " " + eventStatus
	_, err = models.WriteWebEvent(r, "Package", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

func PackageListAll(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve Json format").Error())
		return
	}
	// Parse Json
	var packageJson models.PackageJson
	json.Unmarshal(reqBody, &packageJson)

	// Called List All Package
	packageList, err := models.ListAllPackge(packageJson.Host)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get installed package").Error())
		return
	}
	mapPaginate := models.PaginatePackageList(packageList)
	returnPackageList := models.ReturnPackgeList(mapPaginate, packageJson.Page)

	// Return json
	utils.JSON(w, http.StatusOK, returnPackageList)

}
