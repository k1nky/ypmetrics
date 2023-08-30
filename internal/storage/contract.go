// Пакет storage реализует хранилище метрик
package storage

type storageLogger interface {
	Error(template string, args ...interface{})
}
