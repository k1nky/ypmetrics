// Пакет metric реализует метрики Gauge и Counter
package metric

import (
	"errors"
	"fmt"
	"strconv"
)

// Type тип метрики
type Type string

const (
	TypeGauge   = Type("gauge")
	TypeCounter = Type("counter")
)

// Measure интерфейс для работы с метриками
type Measure interface {
	fmt.Stringer
	// GetName возвращает имя метрики
	GetName() string
	// GetValue возвращает значение метрики
	GetValue() interface{}
	// GetType возвращает тип метрики
	GetType() Type
	// StringValue возвращает значение метрики как строку
	StringValue() string
	// Update обновляет значение метрики
	Update(value interface{}) error
}

// Gauge метрика "измеритель". При обновлении новое значение замещает старое.
type Gauge struct {
	// Имя метрики
	Name string
	// Значение метрики
	Value float64
}

// Counter метрика "счетчик". При обновлении к старому значению добавляется новое.
type Counter struct {
	// Имя метрики
	Name string
	// Значение метрики
	Value int64
}

var (
	ErrInvalidType  = errors.New("invalid type")
	ErrInvalidValue = errors.New("invalid value")
	ErrEmptyName    = errors.New("metric has no name")
)

// IsValid возвращает true, если тип метрики имеет допустимое значение.
func (t Type) IsValid() bool {
	switch t {
	case TypeGauge, TypeCounter:
		return true
	default:
		return false
	}
}

// New создает новую метрики с указаным типом type и именем name.
// Ошибка будет != nil, если указан неверный тип или имя.
func New(typ Type, name string) (Measure, error) {
	if !Type(typ).IsValid() {
		return nil, ErrInvalidType
	}
	if len(name) == 0 {
		return nil, ErrEmptyName
	}
	switch typ {
	case TypeGauge:
		return &Gauge{
			Name: name,
		}, nil
	case TypeCounter:
		return &Counter{
			Name: name,
		}, nil
	}

	return nil, nil
}

// New создает новую метрики с указаным типом type, именем name и значением value.
// Функция аналогична New за исключением, того что иницилизирует ее указанным значением.
func NewWtihValue(typ Type, name string, value interface{}) (Measure, error) {
	m, err := New(typ, name)
	if err != nil {
		return nil, err
	}
	if err = m.Update(value); err != nil {
		return nil, err
	}
	return m, nil
}

func (g Gauge) GetName() string {
	return g.Name
}

func (g Gauge) GetType() Type {
	return TypeGauge
}

func (g Gauge) GetValue() interface{} {
	return g.Value
}

func (g Gauge) String() string {
	return fmt.Sprintf("gauge/%s/%f", g.Name, g.Value)
}

func (g Gauge) StringValue() string {
	return fmt.Sprintf("%g", g.Value)
}

// Update обновляет значение измерителя заменой текущего значения на указанное в value.
// value может быть типа string,uint64,float64, в противном случае будет
// возвращена ошибка.
func (g *Gauge) Update(value interface{}) error {
	var (
		v   float64
		err error
	)
	switch value := value.(type) {
	case string:
		v, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%s: %w allowed types string,uint64,float64", ErrInvalidValue, err)
		}
	case uint64:
		// для простоты рабоы с метриками из пакета runtime
		v = float64(value)
	case float64:
		v = value
	default:
		return ErrInvalidValue
	}
	g.Value = v
	return nil
}

func (c Counter) GetName() string {
	return c.Name
}

func (c Counter) GetType() Type {
	return TypeCounter
}

func (c Counter) GetValue() interface{} {
	return c.Value
}

func (c Counter) String() string {
	return fmt.Sprintf("counter/%s/%d", c.Name, c.Value)
}

func (c Counter) StringValue() string {
	return fmt.Sprintf("%d", c.Value)
}

// Update обновляет значение счетчика, добавляя к текущему значению указанное в value.
// value может быть типа string,int,int64, в противном случае будет
// возвращена ошибка.
func (c *Counter) Update(value interface{}) error {
	var (
		v   int64
		err error
	)
	switch value := value.(type) {
	case string:
		v, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: %w allowed types string,int,int64 ", ErrInvalidValue, err)
		}
	case int:
		v = int64(value)
	case int64:
		v = value
	default:
		return ErrInvalidValue
	}
	c.Value += v
	return nil
}
