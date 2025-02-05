package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/adhocore/gronx"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read request body").Error())
		return
	}

	var task models.Task
	json.Unmarshal(body, &task)
	task.UserId, err = auth.ExtractUserId(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read user id from token").Error())
		return
	}
	template, err := models.GetTemplateFromId(task.TemplateId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("no template with this id exists").Error())
		return
	}
	if task.CronTime != "" {
		gron := gronx.New()
		// check if expr is even valid, returns bool
		if !gron.IsValid(task.CronTime) {
			utils.ERROR(w, http.StatusBadRequest, errors.New("invalid cron time format").Error())
			return
		}
	}
	task.Alert = template.Alert

	if task.CronTime != "" {
		go task.CronRunTask(r)
	} else {
		err = task.RunTask(r)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
		}
	}
}
