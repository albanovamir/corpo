package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileStats содержит статистику по файлу
type FileStats struct {
	Filename    string
	WordCount   int
	SymbolCount int
	Error       error
}

// analyzeFile анализирует один файл и возвращает статистику
func analyzeFile(filename string, results chan<- FileStats, wg *sync.WaitGroup) {
	defer wg.Done()

	stats := FileStats{Filename: filename}

	file, err := os.Open(filename)
	if err != nil {
		stats.Error = err
		results <- stats
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	wordCount := 0
	symbolCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		symbolCount += len(line)

		words := strings.Fields(line)
		wordCount += len(words)
	}

	if err := scanner.Err(); err != nil {
		stats.Error = err
		results <- stats
		return
	}

	stats.WordCount = wordCount
	stats.SymbolCount = symbolCount
	results <- stats
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: program <файл1> [файл2 ...]")
		return
	}

	files := os.Args[1:]
	results := make(chan FileStats, len(files))
	var wg sync.WaitGroup

	// Запускаем анализ каждого файла в отдельной горутине
	for _, file := range files {
		wg.Add(1)
		go analyzeFile(file, results, &wg)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем и выводим результаты
	fmt.Println("Результаты анализа:")
	totalWords := 0
	totalSymbols := 0
	fileNumber := 1

	for result := range results {
		if result.Error != nil {
			fmt.Printf("%d. %s: ошибка - %v\n", fileNumber, result.Filename, result.Error)
		} else {
			fmt.Printf("%d. %s: %d слов, %d символов\n",
				fileNumber,
				filepath.Base(result.Filename),
				result.WordCount,
				result.SymbolCount)

			totalWords += result.WordCount
			totalSymbols += result.SymbolCount
		}
		fileNumber++
	}

	fmt.Printf("\nИтог: %d слов, %d символов.\n", totalWords, totalSymbols)
}

// .\main.exe НовыйТекстовыйДокумент.txt НовыйТекстовыйДокумент1.txt НовыйТекстовыйДокумент2.txt
