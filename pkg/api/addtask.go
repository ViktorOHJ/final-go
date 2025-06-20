package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go1f/pkg/db"
)

// ResponseID представляет собой структуру для ответа с идентификатором задачи в формате JSON.
type ResponseID struct {
	ID string `json:"id"`
}

// ResponseError представляет собой структуру для ответа с ошибкой в формате JSON.
type ResponseError struct {
	Error string `json:"error"`
}

// WriteJSON записывает данные в формате JSON в ответ HTTP.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Ошибка десериализации JSON: %v", err))
	}
}

// WriteError записывает ошибку в формате JSON в ответ HTTP.
func WriteError(w http.ResponseWriter, status int, errMsg string) {
	WriteJSON(w, status, ResponseError{Error: errMsg})
}

// AddTaskHandler обрабатывает HTTP запросы для добавления новой задачи в базу данных.
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	task := db.Task{}

	// 1. Десериализовать JSON запрос в переменную task.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Ошибка чтения тела запроса")
		return
	}
	if err := json.Unmarshal(body, &task); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Ошибка десериализации JSON: %v", err))
		return
	}

	// 2. Проверить, что поле task.Title не пустое.
	if task.Title == "" {
		WriteError(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	// 3. Проверить корректность даты и обработать логику повторения.
	err = checkDate(&task)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Вызвать функцию db.AddTask(task), чтобы добавить задачу в базу данных.
	id, err := db.AddTask(&task)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Ошибка добавления задачи в базу данных: %v", err))
		return
	}

	// 5. Вернуть идентификатор добавленной задачи в виде JSON.
	WriteJSON(w, http.StatusOK, ResponseID{ID: fmt.Sprintf("%d", id)})
}

// checkDate проверяет и корректирует дату задачи в соответствии с правилами.
func checkDate(task *db.Task) error {
	now := time.Now()
	todayFormatted := now.Format("20060102")

	// Если task.Date пустая строка, то присваиваем ему текущую дату.
	if task.Date == "" {
		task.Date = todayFormatted
	}

	// Проверяем, что в task.Date указана корректная дата.
	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return fmt.Errorf("дата представлена в формате, отличном от 20060102: %w", err)
	}

	// Если дата задачи меньше сегодняшнего числа.
	if afterNow(now, t) {
		// Если правило повторения не указано или равно пустой строке, подставляется сегодняшнее число.
		if task.Repeat == "" {
			task.Date = todayFormatted
		} else {
			// При указанном правиле повторения вычисляем следующую дату, которая будет больше сегодняшнего числа.
			nextDateStr, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
			task.Date = nextDateStr
		}
	} else {
		// Если дата задачи не меньше сегодняшнего числа, но repeat указан,
		// все равно проверяем repeat на корректность.
		if task.Repeat != "" {
			_, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
		}
	}
	return nil
}
