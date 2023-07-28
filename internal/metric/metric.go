package metric

import (
	"errors"
	"fmt"
	"strconv"
)

type Type string

const (
	TypeGauge   = Type("gauge")
	TypeCounter = Type("counter")
)

type Measure interface {
	GetName() string
	Update(value interface{}) error
}

type Gauge struct {
	Name  string
	Value float64
}

type Counter struct {
	Name  string
	Value int64
}

var (
	ErrInvalidType  = errors.New("invalid type")
	ErrInvalidValue = errors.New("invalid value")
)

func (t Type) IsValid() bool {
	switch t {
	case TypeGauge, TypeCounter:
		return true
	default:
		return false
	}
}

func New(typ Type, name string) (Measure, error) {
	if !Type(typ).IsValid() {
		return nil, ErrInvalidType
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

func (g *Gauge) GetName() string {
	return g.Name
}

func (g *Gauge) Update(value interface{}) error {
	var (
		v   float64
		err error
	)
	switch value.(type) {
	case string:
		v, err = strconv.ParseFloat(value.(string), 64)
		if err != nil {
			return fmt.Errorf("%s: %w", ErrInvalidValue, err)
		}
	case float64:
		v = value.(float64)
	default:
		return ErrInvalidValue
	}
	g.Value = v
	return nil
}

func (g *Gauge) String() string {
	return fmt.Sprintf("[gauge] %s=%f", g.Name, g.Value)
}

func (c *Counter) GetName() string {
	return c.Name
}

func (c *Counter) Update(value interface{}) error {
	var (
		v   int64
		err error
	)
	switch value.(type) {
	case string:
		v, err = strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return fmt.Errorf("%s: %w", ErrInvalidValue, err)
		}
	case int64:
		v = value.(int64)
	default:
		return ErrInvalidValue
	}
	c.Value += v
	return nil
}

func (c *Counter) String() string {
	return fmt.Sprintf("[counter] %s=%d", c.Name, c.Value)
}
