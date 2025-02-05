package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func SystemInfoGetAllRoute(w http.ResponseWriter, r *http.Request) {

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	sshConnectionList, err := models.GetAllSSHConnectionWithPassword()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve ssh Connection list").Error())
		return
	}
	systemInfoList, err := models.GetAllSysInfo(sshConnectionList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve system info list").Error())
	} else {
		utils.JSON(w, http.StatusOK, systemInfoList)
	}

}

func GetSystemInfoRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to get sshConnection").Error())
		return
	}
	var systemInfo models.SysInfo
	systemInfo, err = models.GetLatestSysInfo(*sshConnection)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("fail to get system info").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
	} else {
		utils.JSON(w, http.StatusOK, systemInfo)
	}

}
