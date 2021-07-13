package routes

import (
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllSSHKey(w http.ResponseWriter, r *http.Request) {
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, err.Error())
		return
	}

	sshKeyList, err := models.GetAllSSHKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}
	utils.JSON(w, http.StatusOK, sshKeyList)
}
