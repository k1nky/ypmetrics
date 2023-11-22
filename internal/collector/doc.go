// Модуль collector включается в себя сборщиков метрик разного типа.
// Каждый сборщик реализует следующий интерфейс:
//
//	  type Collector interface {
//		   Collect(ctx context.Context) (metric.Metrics, error)
//		   Init() error
//	  }
package collector
