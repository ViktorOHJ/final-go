package api

import (
	"net/http"

	"go1f/pkg/db"
)

// TasksResp представляет собой структуру для ответа с задачами в формате JSON.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// GetTasksHandler обрабатывает HTTP запросы для получения списка задач.
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
