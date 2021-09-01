package routes

import (
	"errors"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
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
	var eventStatus string
	if err != nil && err.Error() != "sql: no rows in result set" {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve system info list").Error())
		eventStatus = "failed"
	} else {
		utils.JSON(w, http.StatusOK, systemInfoList)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Get all system info " + eventStatus
	_, err = event.WriteWebEvent(r, "SystemInfo", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}
