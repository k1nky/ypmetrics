// Статический анализатор, который включает в себя:
// - стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
// - всех анализаторов класса SA пакета staticcheck.io;
// - не менее одного анализатора остальных классов пакета staticcheck.io;
// - двух или более любых публичных анализаторов на ваш выбор.
//
// В качестве публичный анализаторов выбраны:
// - https://github.com/gostaticanalysis/sqlrows поиск распространенных ошибок при работе с *sql.Rows;
// - https://github.com/kisielk/errcheck поиск необработанных ошибок.
// Из проверки errcheck исключены файлы тестов. Также исключены сообщения на необработанные ошибки в конструкции defer. Это очень спорный момент (https://github.com/kisielk/errcheck/issues/55), но я решил их исключить, т.к.:
// 1) в рамках данного проекта не вижу смысла обработки ошибок в defer и/или нагромаждения конструкций вида defer `func() { _ = something.Close() }()`
// 2) было интересно понять, можно исклюлчить defer, не внося изменения в errcheck.
package main

import (
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"github.com/k1nky/ypmetrics/osexitanalyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var checks []*analysis.Analyzer

	// проверки staticheck класса simple
	simpleChecks := map[string]bool{
		// условие может быть заменены одним выражением return (https://staticcheck.dev/docs/checks/#S1008)
		"S1008": true,
		// исключаем лишнюю проверку slice на nil перед циклом (https://staticcheck.dev/docs/checks/#S1031)
		"S1031": true,
	}
	// проверки staticheck класса stylecheck
	styleChecks := map[string]bool{
		// проверка формата имен переменных и пакетов (https://staticcheck.dev/docs/checks/#ST1003)
		"ST1003": true,
		// проверка формата сообщения об ошибке (https://staticcheck.dev/docs/checks/#ST1005)
		"ST1005": true,
		// проверка формата имени приемника метода (https://staticcheck.dev/docs/checks/#ST1006)
		"ST1006": true,
		// проверка имен переменных типа time.Duration (https://staticcheck.dev/docs/checks/#ST1011)
		"ST1011": true,
	}
	// проверки staticheck класса quickfix
	quickfixChecks := map[string]bool{
		// возможность замены if/else на switch (https://staticcheck.dev/docs/checks/#QF1003)
		"QF1003": true,
	}

	checks = append(checks,
		// поиск бесплозеных присваиваний
		assign.Analyzer,
		// проверка соответствия используемого шаблона в Printf
		printf.Analyzer,
		// поиск затененных переменных
		shadow.Analyzer,
		// проверка валидности формата тегов полей структур
		structtag.Analyzer,
		// проверка использования прямого вызова os.Exit.
		// Из проверки исключаем файлы, в полном имени которых встречается "go-build" (например, файла из кеша)
		skipFiles(osexitanalyzer.Analyzer, "go-build"),
		// поиск распространенных ошибок при работе с *sql.Rows (https://github.com/gostaticanalysis/sqlrows)
		sqlrows.Analyzer,
		// поиск необработанных ошибок. Из проверки исключаем файлы тестов и сообщения связанные с конструкцией defer.
		skipFiles(skipLines(errcheck.Analyzer, "defer"), "_test.go"),
	)

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		if simpleChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if styleChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if quickfixChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(checks...)
}
