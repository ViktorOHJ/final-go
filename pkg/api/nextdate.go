package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// afterNow возвращает true, если date больше now.
func afterNow(date, now time.Time) bool {
	// Сравниваем только даты, игнорируя время.
	return date.Year() > now.Year() ||
		(date.Year() == now.Year() && date.Month() > now.Month()) ||
		(date.Year() == now.Year() && date.Month() == now.Month() && date.Day() > now.Day())
}

// NextDate вычисляет следующую дату повторения задачи.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("в параметре repeat — пустая строка")
	}

	startDate, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", fmt.Errorf("время в переменной dstart не может быть преобразовано в корректную дату: %w", err)
	}

	parts := strings.Split(repeat, " ")
	currentDate := startDate

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("d: не указан интервал в днях")
		}
		intervalDays, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("d: интервал в днях должен быть числом")
		}
		if intervalDays <= 0 || intervalDays > 400 {
			return "", fmt.Errorf("d %d: превышен максимально допустимый интервал (1-400)", intervalDays)
		}

		// Цикл для сдвига даты на указанное количество дней
		for {
			currentDate = currentDate.AddDate(0, 0, intervalDays)
			if afterNow(currentDate, now) {
				break
			}
		}
	case "y":
		if len(parts) != 1 {
			return "", errors.New("y: этот параметр не требует дополнительных уточнений")
		}
		// Цикл для сдвига даты на год
		for {
			currentDate = currentDate.AddDate(1, 0, 0)
			if afterNow(currentDate, now) {
				break
			}
		}
	default:
		return "", fmt.Errorf("указан неверный формат repeat: %s", repeat)
	}

	return currentDate.Format("20060102"), nil
}
