Дано: Коллега написал функцию `CheckWebsites`, которая проверяет статус сайта для списка URL

```go
package concurency

// WebsiteChecker ...
type WebsiteChecker func(string) bool

// CheckWebsites ...
func CheckWebsites(wc WebsiteChecker, urls []string) map[string]bool {
	results := make(map[string]bool)

	for _, url := range urls {
		results[url] = wc(url)
	}
	return results
}
```

Она возвращает соответсвие для каждого URL отмеченым булевым `true`, для удачного запроса и `fase` для неудачного

Вы должны реализовать `WebsiteChecker` которая принимает URL и возвращает булево значение. Она используется для проверки всех сайтов

Использование внедрения зависимостей позволило протестировать функцию без создания реальных HTTP вызовов, что делает функцию надежной и быстрой. 

Вот тест, который они написали:

```go
package concurency

import (
	"reflect"
	"testing"
)

func mockWebsiteChecker(url string) bool {
	if url == "waat://furhurterwe.degs" {
		return false
	}
	return true
}

func TestCheckWebsites(t *testing.T) {
	websites := []string{
		"http://google.com",
		"http://blog.pypsydave5.com",
		"waat://furhurterwe.degs",
	}
	want := map[string]bool{
		"http://google.com":          true,
		"http://blog.pypsydave5.com": true,
		"waat://furhurterwe.degs":    false,
	}

	got := CheckWebsites(mockWebsiteChecker, websites)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Wanted %v, got %v", want, got)
	}
}
```

Функция расположена в породакшене, и используется для проверки сотен сайтов. Однако коллега начал получать жалобы на медлительность функции и попросил Вас помочь ускорить её

## Напишем тест

```go
func slowStubWebsiteChecker(_ string) bool {
	time.Sleep(20 * time.Millisecond)
	return true
}

func BenchmarkCheckWebsites(b *testing.B) {
	urls := make([]string, 100)

	for i := 0; i < cap(urls); i++ {
		urls[i] = fmt.Sprintf("url #%q", i)
	}
	for i := 0; i < b.N; i++ {
		CheckWebsites(slowStubWebsiteChecker, urls)
	}
}

```
Бенчмарк тестирует `CheckWebsites` используя срез и сотни URL-ов и использует новую псевдореализацию `WebsiteChecker` `slowStubWebsiteChecker` осознано заторможен. Он использует `time.Sleep` для задержки в 20 мс. и возвращает истиное значение.

Когда бенчмарк будет запущен (`go test -bench=.` для Windows PowerShell `go test -bench="."`)

    BenchmarkCheckWebsites-8               1        2043093700 ns/op
    PASS
    ok      github.com/windnow/concurency   2.242s

`CheckWebsites` заняло 2043093700 наносекунд - около двух с четвертью секунд.

Давайте попробуем его ускорить

## Напишем код, достаточный для прохождения

Наконец, мы можем поговорить о конкурентности, в целях получения 'более одной задачи на процесс'. Это то, чем мы реально занимемся ежедневно

Например, этим утром я заварил чашку чая. Я включил чайник, и затем, пока он закипал, достал молоко из холдильника, достал чай из буфета, нашел любимую кружку, положил пакетик в кружку и затем, когда чайник закипел, залил его кипятком.

Я не ждал, включив чайник, пока он закипит, просто уставившись на него и после закипания не сделал все остальное.

Если вы понимаете, как приготовить чай быстрее, вы так же поймёте, как мы ускорим `CheckWebsites`. Вместо того, что бы ждать ответа веб-сайта после отправки запроса, мы скажем компьютеру, что нужно отправить следующий запрос, пока он ожидает.

Обычно в программировании когда вызывается некоторая функция, мы ожидаем от нее возвращаемого значения (если же функция не возвращает ничего, то мы все равно ждем её завершения) мы говорим, что эта операция блокирует - т.е. заставляет нас ждать своего завершения. Операция, которая не блокирует запускается в отдельном процессе, называемом *горутиной*.

Что бы сказать Go запустить новую горутину мы передаем вызов функции оператору с ключевы словом `go`: `go doSomething()`.

