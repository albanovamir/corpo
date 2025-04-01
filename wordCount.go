package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type fileStr struct {
	filePath string
}

func main() {
	for {
		fmt.Println("")

		// Запрашивание у пользователя путь к файлу
		fmt.Print("Введите путь к текстовому файлу: ")
		var filePath = fileStr{}
		fmt.Scanln(&filePath.filePath)
		if filePath.filePath == "/exit/" {
			break
		}

		// Запрашивание у пользователя слово для поиска
		fmt.Print("Введите слово для поиска: ")
		var searchWord string
		fmt.Scanln(&searchWord)
		if searchWord == "/exit/" {
			break
		}

		// Открытие файла
		file, err := os.Open(filePath.filePath)
		if err != nil {
			fmt.Println("Ошибка при открытии файла:", err)
		}
		defer file.Close()

		// Инициализация счетчика
		var wordCount, searchWordCount int

		// Чтение файла построчно
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// избавление от знаков препинания
			line = strings.Map(isPunctuation, line)
			// Разбитие строки на слова
			words := strings.Fields(line)
			wordCount += len(words)

			// Подсчет количества искомого слова
			for _, word := range words {
				if strings.EqualFold(word, searchWord) {
					searchWordCount++
				}
			}
		}

		// Проверка на наличие ошибок при чтении файла
		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при чтении файла:", err)
		}

		// Вывод результатов
		fmt.Printf("Общее количество слов в файле: %d\n", wordCount)
		fmt.Printf("Количество повторений слова '%s': %d\n", searchWord, searchWordCount)
	}

	fmt.Println("Завершение программы...")
}

// функция для поиска знаков препинания
func isPunctuation(r rune) rune {
	if unicode.IsPunct(r) {
		return -1
	}
	return r
}

// путь до файла:
// C:\Users\amira\Folder\Desktop1\Studying\MIREA\2\industrialSysProgramming\practice1\НовыйТекстовыйДокумент.txt

// xml докментация
// самари коментм
// солид
// стэк хип
//http сокеты tcp ip
