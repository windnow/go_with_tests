Context 
=======
[Оригинал](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/context)

Часто запуск программное обеспечение запускает достаточно долгоие ресурсоемкие процессы (обычно в горутинах). Если действие, вызвавшее их будет по какой то причине отменено, или не выполнено, то необходимо последовательно остановить эти процессы в приложении.

Если не управлять этим, то go приложение может вызвать проблемы с отладкой производительности.

В этой главе мы будем использовать пакет `context` для облегчения управления долго работающими процессами.

Мы начнем с классического примера web сервера, который при необходимости запускает потенциально длительный процесс для извлечения некоторых данных для возврата в ответ.

Мы реализуем сценарий, в котором пользователь отменяет запрос до того, как данные могут быть получены, и убедимся что процесс был остановлен

Я реализовал некоторый статовый код

```go
func Server(store Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request){
        fmt.Fprint(w, store.Fetch())
    }
}
```

Функция `Server` принимает `Store` и возвращает `http.HandlerFunc`. `Store` определено следующим образом:

```go
type Store interface {
    Fetch() string
}
```

Возвращаемая функция вызывает метод `Fetch` обЪекта `store` и записывает его в ответ (`response`)

У нас есть соответствующая заглушка для `Store`, которую мы используем в тесте

```go
type StubStore struct {
	response string
}

func (s *StubStore) Fetch() string {
	return s.response
}

func TestHandler(t *testing.T) {
	data := "hello, world"
	srv := Server(&StubStore{data})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	srv.ServeHTTP(response, request)

	if response.Body.String() != data {
		t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
	}
}
```
Теперь, когда у нас есть удачный сценарий, необходимо реализовать более реализстичный сценарий, в котором `Store` не может завершить `Fetch` до того, как пользователь отменит запрос.

## Сначал напишем тест
Нашему обработчику потребуется спосом сообщить хранилищу о отмене работы, потому обновим интерфейс 

```go
type Store interface {
	Fetch() string
	Cancel()
}
```

Нужно настроить шпиона т.о. что бы данные возвращались через некоторое время, в течении которого может произойти отмена. Мы так же переименуем его в `SpyStore`, т.к. теперь мы следим за тем, каким образом он вызывается. Нужно добавить метод `Cancel`, реализуя интерфейс `Store`

```go
type SpyStore struct {
	response  string
	cancelled bool
}

func (s *SpyStore) Fetch() string {
    time.Sleep(100 * time.Millesecond)
	return s.response
}

func (s *SpyStore) Cancel() {
	s.cancelled = true
}
```

Давайте добавим тест, в котором мы отмениз запрос до истечения 100 миллисекунд, и проверим хранилище на факт отмены

```go
	t.Run("tells store to cancel work if request is cancelled", func(t *testing.T) {
		store := &SpyStore{response: data}
		srv := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		cancellingCtx, cancel := context.WithCancel(request.Context())
		time.AfterFunc(5*time.Millisecond, cancel)
		request = request.WithContext(cancellingCtx)

		response := httptest.NewRecorder()
		srv.ServeHTTP(response, request)
		if !store.cancelled {
			t.Errorf("store was not told to cancel")
		}
	})
```

Взято с [блога Go: Context](https://blog.golang.org/context)

    Пакет context предоставляет функции для получения новых значний контекста из существующих. Эти значение образуют дерево: при отмене контекста производные от него контексты так же отменяются

Важно получение контекста, что бы отмены распространялись по всему стеку вызова для данного запроса.

Мы получаем наш контекст `cancellingCtx` из запроса `request`, который тек же возвращает функцию отмены `cancel`. Затем запускается планировщик `time.AfterFunc`, который вызывает эту функцию через 5 миллисекунд. Наконец, мы используем наш новый контекст в запросе вызываю `request.WithContext`

## Попробуйте запустить тест

Тест, как и ожидалось, провалится

## Напишите минимум кода для его прохождения

Помните о дисциплине TDD. Напишите минимальное количесвто кода, что бы пройти наш тест
```go
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.Cancel()
		fmt.Fprint(w, store.Fetch())
	}
}
```
Это позволяет пройти тестам успешна, но делает это не очень хорошо. Разумеется, мы не должны отменять наше хранилище `Store` при *каждом запросе*

Соблюдение дисциплины осветило изъян нашего теста, и это хорошо!

Теперь нам нужно обновить наши тесты, что бы проверить, не является ли он отмененным.

```go
t.Run("returns data from store", func(t *testing.T) {
		store := &SpyStore{response: data}
		srv := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		if response.Body.String() != data {
			t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
		}

		if store.cancelled {
			t.Error("it sould not have cancelled the store")
		}
    })
```
    
Запустите оба теста и один из них должен завершиться неудачей, и теперь мы вынуждены сделать более разумную реализацию

```go
 func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data := make(chan string, 1)

		go func() {
			data <- store.Fetch()
		}()
		select {
		case d := <-data:
			fmt.Fprint(w, d)
		case <-ctx.Done():
			store.Cancel()
		}
	}
}
```
Что же мы сдесь сделали?

`context` имеет метод `Done()`, возвращающий канал, получающий сигнал при завершении или отмене контекста. Нам нужно слушать данный сигнал и вызываеть `store.Cancel` если получим его, однако он должен быть проигнорирован, если `Fetch` нашего хранилища завершится раньше.

Для реализации этого мы запускаем `Fetch` в горутине, которая будет писать результат в новый канал `data`. Затем используем `select` для эффективной связки двух асинхронных процессов, что бы либо записать в `response`, либо вызвать `Cancel`