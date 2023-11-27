package handler

import (
	"bytes"
	"net/http"
)

//lint:file-ignore all Ignore all unused code, it's generated

func ExampleHandler_AllMetrics() {
	resp, _ := http.Get("http://localhost:8080/")
	resp.Body.Close()
}

func ExampleHandler_Value() {
	// получаем значение счетчика с именем counter0
	resp, _ := http.Get("http://localhost:8080/counter/counter0")
	resp.Body.Close()

	// получаем значение измерителя с именем gauge0
	resp, _ = http.Get("http://localhost:8080/gauge/gauge0")
	resp.Body.Close()
}

func ExampleHandler_Update() {
	// отправить метрику типа счетчик с именем counter0 и значением 10
	resp, _ := http.Post("http://localhost:8080/counter/counter0/10", "text/plain", nil)
	resp.Body.Close()

	// отправить метрику типа измеритель с именем gauge0 и значением 10.1
	resp, _ = http.Post("http://localhost:8080/gauge/gauge0/10.1", "text/plain", nil)
	resp.Body.Close()
}

func ExampleHandler_ValueJSON() {
	// получаем значение счетчика с именем counter0
	payload := bytes.NewBufferString(`
		{
			"id": "counter0",
			"type":"counter"
		}
	`)
	resp, _ := http.Post("http://localhost:8080/value/", "application/json", payload)
	resp.Body.Close()

	// получаем значение измерителя с именем gauge0
	payload = bytes.NewBufferString(`
		{
			"id": "gauge0",
			"type":"gauge"
		}
	`)
	resp, _ = http.Post("http://localhost:8080/value/", "application/json", payload)
	resp.Body.Close()

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
	resp, _ := http.Post("http://localhost:8080/update/", "application/json", payload)
	resp.Body.Close()

	// отправить метрику типа измеритель с именем gauge0 и значением 10.1
	payload = bytes.NewBufferString(`
		{
			"id": "gauge0",
			"type":"gauge",
			"value": 10.1
		}
	`)
	resp, _ = http.Post("http://localhost:8080/update/", "application/json", payload)
	resp.Body.Close()
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
	resp, _ := http.Post("http://localhost:8080/updates/", "application/json", payload)
	resp.Body.Close()

}

func ExampleHandler_Ping() {
	resp, _ := http.Get("http://localhost:8080/ping")
	resp.Body.Close()
}
