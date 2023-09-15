package retrier

import "time"

type ShouldRetry func(err error) bool

// Retrier контроллер повторного выполнения
//
//	for r := retrier.New(myShouldRetry); r.Next(err); {
//		err = doSomething()
//	}
type Retrier struct {
	// счетчик попыток
	attempt int
	// задержка между попытками
	retries []time.Duration
	// функция, которая проверяет ошибку на Retriable
	shouldRetry ShouldRetry
}

// Возвращает новый контроллер повторного выполнения
func New() *Retrier {
	return &Retrier{
		attempt: 0,
		// по умолчанию 1 сек, 3 сек, 5 сек
		retries: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (r *Retrier) Init(shouldRetry func(error) bool) {
	r.attempt = 0
	r.shouldRetry = shouldRetry
}

// Возвращает true если следует повторить действие. False - нет ошибки или попытки кончались.
func (r *Retrier) Next(err error) bool {
	// всегда увеличиваем счетчик попыток
	defer func() {
		r.attempt += 1
	}()
	if r.attempt == 0 {
		// еще не было попыток
		return true
	}
	if err == nil {
		return false
	}
	if (r.attempt <= len(r.retries)) && r.shouldRetry(err) {
		time.Sleep(r.retries[r.attempt-1])
		return true
	}
	return false
}

// AlwaysRetry удобно использовать, если планируем повторять при любой ошибке.
func AlwaysRetry(err error) bool {
	return true
}
