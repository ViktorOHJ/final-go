package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func GetTask(id string) (*Task, error) {
	if DB == nil {
		return nil, errors.New("db.DB is nil: database connection not initialized")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("некорректный формат ID: %w", err) // Ошибка парсинга
	}
	row := DB.QueryRow(query, idInt) // Используем int64 для запроса к числовому ID в БД

	var task Task
	var dbID int64 // Всегда сканируем числовой ID из БД в int64

	if err := row.Scan(&dbID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача с ID %s не найдена", id) // <-- Используем ErrTaskNotFound
		}
		return nil, fmt.Errorf("ошибка при сканировании строки задачи: %w", err)
	}
	// Преобразуем int64 ID из БД в string для поля Task.ID
	task.ID = strconv.FormatInt(dbID, 10)

	return &task, nil
}

func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи в БД: %w", err)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID последней вставленной записи: %w", err)
	}
	task.ID = strconv.FormatInt(lastID, 10)
	return lastID, nil
}
func Tasks(w http.ResponseWriter, limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к БД: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении строк: %w", err)
	}
	if len(tasks) == 0 {
		tasks = []*Task{} // Возвращаем пустой срез, если нет задач
		return tasks, nil
	}
	return tasks, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	// Проверяем, что обновление затронуло хотя бы одну строку
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("некорректный формат ID: %w", err) // Ошибка парсинга
	}
	res, err := DB.Exec(query, idInt)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи из БД: %w", err)
	}
	// Проверяем, что удаление затронуло хотя бы одну строку
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача с ID %s не найдена для удаления", id) // <-- Используем ErrTaskNotFound
	}
	return nil
}
