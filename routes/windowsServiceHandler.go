package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsServiceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	serviceList, err := sshConnection.GetServiceList()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, serviceList)
}

func ChangeWindowsServiceState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	serviceName := vars["service_name"]
	serviceState := vars["service_state"]

	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = sshConnection.ChangeWindowsServiceState(serviceName, serviceState)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}
