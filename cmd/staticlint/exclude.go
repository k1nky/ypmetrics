package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// skipFiles исключает из отчета анализатора сообщения на файл, содержащий в имени строку pattern.
func skipFiles(a *analysis.Analyzer, pattern string) *analysis.Analyzer {
	run := a.Run
	// дополняем оригинальную функцию Run анализатора своей проверкой
	a.Run = func(pass *analysis.Pass) (interface{}, error) {
		// аналогично поступаем с методом Report
		report := pass.Report
		pass.Report = func(d analysis.Diagnostic) {
			file := pass.Fset.File(d.Pos)
			if file != nil && !strings.Contains(file.Name(), pattern) {
				report(d)
			}
		}
		return run(pass)
	}
	return a
}

// skipLines исключает из отчета анализатора сообщения на строку, содержащий в себе строку pattern.
func skipLines(a *analysis.Analyzer, pattern string) *analysis.Analyzer {
	run := a.Run
	a.Run = func(pass *analysis.Pass) (interface{}, error) {
		report := pass.Report
		pass.Report = func(d analysis.Diagnostic) {
			// получаем файл, на который ссылается отчет
			file := pass.Fset.File(d.Pos)
			// получаем исходную строку
			// для этого читаем нужную строку из соответсвующего файла
			if f, err := os.Open(file.Name()); err == nil {
				// определяем смещение на начало нужной строки:
				//  (смещение от начала файла на символ, на который указывает анализатор) - (смещение от начала строки, на который указывает анализатор) + 1 (иначе будет конец предыдущей строки)
				offset := file.Offset(d.Pos) - file.Position(d.Pos).Column + 1
				if line, err := readLine(f, offset); err == nil {
					if strings.Contains(line, pattern) {
						return
					}
				}
			}
			report(d)
		}
		return run(pass)
	}
	return a
}

// readLine читает строку из r, начинающую с offset.
func readLine(r io.ReadSeeker, offset int) (string, error) {

	rd := bufio.NewReader(r)
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return "", err
	}
	line, _, err := rd.ReadLine()
	return string(line), err
}
