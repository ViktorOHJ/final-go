package api

import (
	"go1f/pkg/db"
	"net/http"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(w, 50) // в параметре максимальное количество записей
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Ошибка получения задач: "+err.Error())
		return
	}
	WriteJSON(w, http.StatusAccepted, TasksResp{
		Tasks: tasks,
	})
}
