package handler

import (
	"bytes"
	"net/http"
)

func ExampleHandler_AllMetrics() {
	http.Get("http://localhost:8080/")
}

func ExampleHandler_Value() {
	// получаем значение счетчика с именем counter0
	http.Get("http://localhost:8080/counter/counter0")

	// получаем значение измерителя с именем gauge0
	http.Get("http://localhost:8080/gauge/gauge0")
}

func ExampleHandler_Update() {
	// отправить метрику типа счетчик с именем counter0 и значением 10
	http.Post("http://localhost:8080/counter/counter0/10", "text/plain", nil)

	// отправить метрику типа измеритель с именем gauge0 и значением 10.1
	http.Post("http://localhost:8080/gauge/gauge0/10.1", "text/plain", nil)
}

func ExampleHandler_ValueJSON() {
	// получаем значение счетчика с именем counter0
	payload := bytes.NewBufferString(`
		{
			"id": "counter0",
			"type":"counter"
		}
	`)
	http.Post("http://localhost:8080/value/", "application/json", payload)

	// получаем значение измерителя с именем gauge0
	payload = bytes.NewBufferString(`
		{
			"id": "gauge0",
			"type":"gauge"
		}
	`)
	http.Post("http://localhost:8080/value/", "application/json", payload)
}

func ExampleHandler_UpdateJSON() {
	// отправить метрику типа счетчик с именем counter0 и значением 10
	payload := bytes.NewBufferString(`
		{
			"id": "counter0",
			"type":"counter",
			"delta": 10
		}
	`)
	http.Post("http://localhost:8080/update/", "application/json", payload)

	// отправить метрику типа измеритель с именем gauge0 и значением 10.1
	payload = bytes.NewBufferString(`
		{
			"id": "gauge0",
			"type":"gauge",
			"value": 10.1
		}
	`)
	http.Post("http://localhost:8080/update/", "application/json", payload)
}

func ExampleHandler_UpdatesJSON() {
	// отправить несколько метрик
	payload := bytes.NewBufferString(`
		[
			{
				"id": "counter0",
				"type":"counter",
				"delta": 10
			},
			{
				"id": "counter1",
				"type":"counter",
				"delta": 1
			},
			{
				"id": "gauge0",
				"type":"gauge",
				"delta": 1.1
			}
		]
	`)
	http.Post("http://localhost:8080/updates/", "application/json", payload)
}

func ExampleHandler_Ping() {
	http.Get("http://localhost:8080/ping")
}
