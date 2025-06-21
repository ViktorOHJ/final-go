package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go1f/pkg/db"
)

func Init() {
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/nextdate", HandleNextDate)
	http.HandleFunc("/api/tasks", GetTasksHandler)
	http.HandleFunc("/api/task/done", DoneHandler)
	log.Println("Обработчики зарегистрированы.")
}

// TaskHandler обрабатывает HTTP запросы для управления задачами.
// В зависимости от метода запроса (POST, GET, PUT, DELETE) вызывает соответствующие обработчики.
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Получен запрос по пути: %s, метод: %s\n", r.URL.Path, r.Method)
	switch r.Method {
	case http.MethodPost:
		log.Println("Вызов api.AddTaskHandler")
		AddTaskHandler(w, r)
	case http.MethodGet:
		log.Println("Вызов api.GetTask")
		GetTaskByIDHandler(w, r)
	case http.MethodPut:
		log.Println("Вызов api.UpdateTaskHandler")
		UpdateTaskHandler(w, r)
	case http.MethodDelete:
		log.Println("Вызов api.DeleteTaskHandler")
		DeleteTaskHandler(w, r)
	}
}

// DeleteTaskHandler обрабатывает HTTP запросы для удаления задачи по ID.
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	log.Printf("Получен запрос на удаление задачи с ID: %s\n", taskID)

	if taskID == "" {
		WriteError(w, http.StatusBadRequest, "Не указан ID задачи")
		return
	}

	err := db.DeleteTask(taskID)
	if err != nil {
		log.Printf("Ошибка удаления задачи с ID %s: %v\n", taskID, err)
		WriteError(w, http.StatusInternalServerError, "Ошибка удаления задачи: "+err.Error())
		return
	}

	log.Printf("Задача с ID %s успешно удалена.\n", taskID)
	WriteJSON(w, http.StatusOK, struct{}{}) // Отправляем пустой JSON в ответе
}

// DoneHandler обрабатывает HTTP запросы для завершения задачи по ID.
func DoneHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	log.Printf("Получен запрос на завершение задачи с ID: %s\n", taskID)

	if taskID == "" {
		WriteError(w, http.StatusBadRequest, "Не указан ID задачи")
		return
	}

	// 1. Получаем задачу
	task, err := db.GetTask(taskID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Задача не найдена")
		log.Printf("Ошибка получения задачи (внутренняя): %v\n", err)
		WriteError(w, http.StatusInternalServerError, "Ошибка получения задачи: "+err.Error())
		return
	}
	if task.Repeat == "" {
		// 2. Если задача не повторяется, просто удаляем её
		err = db.DeleteTask(taskID)
		if err != nil {
			log.Printf("Ошибка удаления задачи с ID %s: %v\n", taskID, err)
			WriteError(w, http.StatusInternalServerError, "Ошибка удаления задачи: "+err.Error())
			return
		}
		log.Printf("Задача с ID %s успешно удалена (не повторяется).\n", taskID)
	} else {
		// 3. Если задача повторяется, вычисляем следующую дату и обновляем её
		date, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Printf("Ошибка при вычислении следующей даты для задачи ID %s: %v\n", taskID, err)
			WriteError(w, http.StatusBadRequest, "Ошибка при вычислении следующей даты: "+err.Error())
			return
		}
		// Обновляем дату задачи
		task.Date = date
		err = db.UpdateTask(task)
		if err != nil {
			log.Printf("Ошибка обновления задачи с ID %s: %v\n", taskID, err)
			WriteError(w, http.StatusInternalServerError, "Ошибка обновления задачи: "+err.Error())
			return
		}
		log.Printf("Задача с ID %s успешно обновлена до следующей даты: %s\n", taskID, task.Date)
	}

	// 4. Отправляем финальный успешный ответ
	log.Printf("Операция завершения задачи с ID %s успешно выполнена.\n", taskID)
	WriteJSON(w, http.StatusOK, struct{}{})
}

// UpdateTaskHandler обрабатывает HTTP запросы для обновления задачи.
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// 2. Проверить, что поле task.ID не пустое.
	if task.ID == "" {
		WriteError(w, http.StatusBadRequest, "Не указан ID задачи")
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
		WriteError(w, http.StatusBadRequest, "Некорректная дата или повторение: "+err.Error())
		return
	}

	// 4. Обновить задачу в базе данных.
	err = db.UpdateTask(&task)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Ошибка обновления задачи: "+err.Error())
		return
	}

	log.Printf("Задача с ID %s успешно обновлена: %+v\n", task.ID, task)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ResponseID{ID: task.ID}); err != nil {
		WriteError(w, http.StatusInternalServerError, "Ошибка сериализации ответа в JSON: "+err.Error())
		return
	}
	log.Printf("Ответ отправлен для обновленной задачи с ID %s\n", task.ID)
}

// GetTaskByIDHandler обрабатывает HTTP запросы для получения задачи по ID.
func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	log.Printf("Получен запрос на получение задачи с ID: %s\n", taskID)

	if taskID == "" {
		WriteError(w, http.StatusBadRequest, "Не указан ID задачи")
		return
	}

	task, err := db.GetTask(taskID)

	if err != nil {

		WriteError(w, http.StatusNotFound, "Задача не найдена")
		return
	}
	log.Printf("Задача с ID %s успешно получена: %+v\n", taskID, task)
	WriteJSON(w, http.StatusOK, task)
	log.Printf("Ответ отправлен для задачи с ID %s\n", taskID)
}

// HandleNextDate обрабатывает HTTP запросы для вычисления следующей даты задачи.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	queryNow := r.URL.Query().Get("now")
	queryDate := r.URL.Query().Get("date")
	queryRepeat := r.URL.Query().Get("repeat")

	var now time.Time
	var err error

	if queryNow == "" {
		now = time.Now() // Если параметр 'now' не указан, берем текущую дату
	} else {
		now, err = time.Parse("20060102", queryNow)
		if err != nil {
			http.Error(w, "Некорректный формат даты в параметре 'now'. Ожидается 20060102.", http.StatusBadRequest)
			return
		}
	}

	if queryDate == "" {
		http.Error(w, "Параметр 'date' обязателен.", http.StatusBadRequest)
		return
	}

	if queryRepeat == "" {
		http.Error(w, "Параметр 'repeat' обязателен.", http.StatusBadRequest)
		return
	}

	nextDateStr, err := NextDate(now, queryDate, queryRepeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDateStr))
}
