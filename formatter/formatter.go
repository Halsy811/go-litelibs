package formatter

import (
	"fmt"
	"sort"
	"strings"
)

/*
Форматирование в виде листа (Max 2 уровня вложенности)

ПРИМЕР ВЫВОДА:

	key1: 	value1
		  	value2
		  	value3
	key2: 	value1
		  	value2
*/
func FormatList1(data *map[string][]string) {

}

// Форматированный вывод коллекции в виде list
func FormatAsNestedList(data map[string][]string) {
	fmt.Printf("\n")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 1. Находим самую длинную строку среди ключей для выравнивания
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	var sb strings.Builder

	for _, key := range keys {
		values := data[key]
		if len(values) == 0 {
			sb.WriteString(fmt.Sprintf("%-*s: <empty>\n", maxKeyLen, key))
			continue
		}

		// Разбиваем первое значение по переносам строк
		lines := strings.Split(values[0], "\n")
		for j, line := range lines {
			line = strings.ReplaceAll(line, "\r", "")
			if j == 0 {
				// Первая строка: Ключ + Значение с выравниванием
				sb.WriteString(fmt.Sprintf("%-*s: %s\n", maxKeyLen, key, line))
			} else {
				// Остальные строки многострочного значения с отступом под значение
				// Отступ = длина самого длинного ключа + 2 символа (ключ + ": ")
				sb.WriteString(fmt.Sprintf("%*s%s\n", maxKeyLen+2, "", line))
			}
		}

		// Остальные значения массива (если их несколько)
		for i := 1; i < len(values); i++ {
			lines := strings.Split(values[i], "\n")
			for _, line := range lines {
				line = strings.ReplaceAll(line, "\r", "")
				// Тот же отступ
				sb.WriteString(fmt.Sprintf("%*s%s\n", maxKeyLen+2, "", line))
			}
		}
	}

	fmt.Print(sb.String())
}

/*
FormatAnyAsNestedList форматирует произвольный JSON в виде выровненного списка.
Принимает результат json.Unmarshal (map[string ]any, []any или примитив).

ПРИМЕР ВЫВОДА:

	config     :
		endpoints:
				https://api.example.com/v1
				https://backup.example.com/v1
		metadata :
				notes:
						notes:
						owner: team-platform
				owner: team-platform
		retries  : 3
		timeout  : 30

ПРИМЕНЕНИЕ:

var data any

	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
	    panic(err)
	}

	FormatAnyAsNestedList(data)
*/
func FormatAnyAsNestedList(data any) {
	fmt.Printf("\n")
	sb := &strings.Builder{}
	formatValue(sb, data, 0)
	fmt.Print(sb.String())
}

/*
FormatTable выводит данные в виде выровненной таблицы.
data: слайс строк-записей.
fields: список ключей для вывода. Если nil или пустой — выводятся ВСЕ поля
(порядок определяется по первой записи, затем сортируется).

ПРИМЕР ВЫВОДА:

	name          status      cpu
	------------  ----------  ---
	api-gateway   running     12%
	auth-service  degraded    87%
				  restarting  55%

ПРИМЕНЕНИЕ:

	data := []map[string]string{
			{"name": "api-gateway", "status": "running", "cpu": "12%", "memory": "256Mi"},
			{"name": "auth-service", "status": "degraded\nrestarting", "cpu": "87%\n55%", "memory": "1.2Gi"},

// Вариант 1: Только выбранные поля в заданном порядке

	FormatSliceAsTable(data, []string{"name", "status", "cpu"})

// Вариант 2: Все поля (автоматически, отсортировано)

	FormatSliceAsTable(data, nil)
*/
func FormatSliceAsTable(data []map[string]string, fields []string) {
	fmt.Printf("\n")

	if len(data) == 0 {
		fmt.Println("<no data>")
		return
	}

	// --- 1. Определяем колонки ---
	headers := resolveHeaders(data, fields)
	if len(headers) == 0 {
		fmt.Println("<no fields to display>")
		return
	}

	// --- 2. Вычисляем максимальную ширину каждой колонки ---
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = visibleLen(h)
	}
	for _, row := range data {
		for i, h := range headers {
			val := row[h]
			// Учитываем самую длинную строку в многострочном значении
			for _, line := range strings.Split(val, "\n") {
				line = strings.ReplaceAll(line, "\r", "")
				if l := visibleLen(line); l > colWidths[i] {
					colWidths[i] = l
				}
			}
		}
	}

	// --- 3. Строим вывод ---
	var sb strings.Builder

	// Заголовок
	for i, h := range headers {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(fmt.Sprintf("%-*s", colWidths[i], h))
	}
	sb.WriteByte('\n')

	// Разделитель
	for i, w := range colWidths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteByte('\n')

	// Строки данных
	for _, row := range data {
		writeTableRow(&sb, row, headers, colWidths)
	}

	fmt.Print(sb.String())
}

// ==============================================
// =============     internal     ===============
// ==============================================

// formatValue рекурсивно обрабатывает значение с заданным базовым отступом
func formatValue(sb *strings.Builder, data any, baseIndent int) {
	switch v := data.(type) {
	case map[string]any:
		formatMap(sb, v, baseIndent)
	case []any:
		// Массив без ключа: выводим каждый элемент с текущим отступом
		for _, item := range v {
			writeIndentedLine(sb, baseIndent, "- ")
			formatValue(sb, item, baseIndent+2)
		}
	default:
		// Примитив (string, float64, bool, nil)
		writeMultilineValue(sb, baseIndent, fmt.Sprintf("%v", v))
	}
}

// formatMap реализует вашу логику выравнивания для map[string]any
func formatMap(sb *strings.Builder, m map[string]any, baseIndent int) {
	if len(m) == 0 {
		writeIndentedLine(sb, baseIndent, "<empty object>")
		return
	}

	// Сортировка ключей
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Поиск максимальной длины ключа НА ТЕКУЩЕМ УРОВНЕ
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	valueIndent := baseIndent + maxKeyLen + 2 // отступ под значение

	for _, key := range keys {
		val := m[key]

		switch child := val.(type) {
		case map[string]any:
			// Вложенный объект: ключ на одной строке, содержимое со сдвигом
			writeIndentedLine(sb, baseIndent, fmt.Sprintf("%-*s:", maxKeyLen, key))
			formatMap(sb, child, valueIndent)

		case []any:
			// Массив: первое значение рядом с ключом, остальные под ним
			if len(child) == 0 {
				writeIndentedLine(sb, baseIndent, fmt.Sprintf("%-*s: <empty>", maxKeyLen, key))
				continue
			}
			for i, item := range child {
				prefix := fmt.Sprintf("%-*s: ", maxKeyLen, key)
				if i > 0 {
					prefix = ""
				}
				writeIndentedLine(sb, baseIndent, prefix)
				formatValue(sb, item, valueIndent)
			}

		default:
			// Примитивное значение (включая многострочные строки)
			strVal := fmt.Sprintf("%v", val)
			lines := strings.Split(strings.ReplaceAll(strVal, "\r", ""), "\n")
			for j, line := range lines {
				if j == 0 {
					writeIndentedLine(sb, baseIndent, fmt.Sprintf("%-*s: %s", maxKeyLen, key, line))
				} else {
					writeIndentedLine(sb, valueIndent, line)
				}
			}
		}
	}
}

// writeIndentedLine записывает строку с абсолютным отступом
func writeIndentedLine(sb *strings.Builder, indent int, text string) {
	if indent > 0 {
		sb.WriteString(strings.Repeat(" ", indent))
	}
	sb.WriteString(text)
	sb.WriteByte('\n')
}

// writeMultilineValue записывает примитив с поддержкой переносов строк
func writeMultilineValue(sb *strings.Builder, indent int, text string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	for _, line := range lines {
		writeIndentedLine(sb, indent, line)
	}
}

// resolveHeaders определяет итоговый список и порядок колонок.
func resolveHeaders(data []map[string]string, fields []string) []string {
	if len(fields) > 0 {
		// Фильтруем только те, что реально есть хотя бы в одной записи
		existing := make(map[string]bool)
		for _, row := range data {
			for k := range row {
				existing[k] = true
			}
		}
		result := make([]string, 0, len(fields))
		for _, f := range fields {
			if existing[f] {
				result = append(result, f)
			}
		}
		return result
	}

	// Все поля: собираем из всех записей, сортируем для детерминизма
	seen := make(map[string]bool)
	var all []string
	for _, row := range data {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				all = append(all, k)
			}
		}
	}
	sort.Strings(all)
	return all
}

// writeTableRow записывает одну строку таблицы с поддержкой многострочных ячеек.
func writeTableRow(sb *strings.Builder, row map[string]string, headers []string, widths []int) {
	// Разбиваем каждую ячейку на строки
	cellLines := make([][]string, len(headers))
	maxLines := 1
	for i, h := range headers {
		raw := strings.ReplaceAll(row[h], "\r", "")
		lines := strings.Split(raw, "\n")
		cellLines[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Выводим построчно
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		for colIdx := range headers {
			if colIdx > 0 {
				sb.WriteString("  ")
			}
			text := ""
			if lineIdx < len(cellLines[colIdx]) {
				text = cellLines[colIdx][lineIdx]
			}
			sb.WriteString(fmt.Sprintf("%-*s", widths[colIdx], text))
		}
		sb.WriteByte('\n')
	}
}

// visibleLen возвращает видимую длину строки (без учёта ANSI-escape кодов).
// Для простоты здесь используется обычная длина; при необходимости
// можно подключить github.com/mattn/go-runewidth для CJK/emoji.
func visibleLen(s string) int {
	return len(s)
}
