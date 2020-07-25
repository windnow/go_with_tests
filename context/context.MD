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

## Рефакторинг

Мы можем немного реорганизовать код, перенеся методы сравнения в наблюдатель

```go
func (s *SpyStore) assertWasCancelled() {
	s.t.Helper()
	if !s.cancelled {
		s.t.Errorf("store was not told cancel")
	}
}

func (s *SpyStore) assertWasNotCancelled() {
	s.t.Helper()
	if s.cancelled {
		s.t.Errorf("store was told to cancel")
	}
}
```

Не забудьте передать `*testing.T` при создании наблюдателя.

```go
func TestHandler(t *testing.T) {
	data := "hello, world"

	t.Run("returns data from store", func(t *testing.T) {
		store := &SpyStore{response: data, t: t}
		srv := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		if response.Body.String() != data {
			t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
		}

		store.assertWasNotCancelled()
	})
	t.Run("tells store to cancel work if request is cancelled", func(t *testing.T) {
		store := &SpyStore{response: data, t: t}
		srv := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		cancellingCtx, cancel := context.WithCancel(request.Context())
		time.AfterFunc(5*time.Millisecond, cancel)

		request = request.WithContext(cancellingCtx)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		store.assertWasCancelled()
	})
}
```

Такой подход возможен, но характерен ли он?

Имеет ли смысл для нашего сервера иметь дело с ручной отменой `Store`? Что, если `Store` так же может зависеть от другого медленного процесса? Мы должны убедиться что `Store.Cancel` корректно распространяется на все зависимости.

Одним из основных моментов `context` заключается в последовательном распространении отмены

[Из документации go](https://golang.org/pkg/context/)
	
	Входящий запрос к серверу должен создавать контекст, исходящие же вызовы на
	серверы должны принимать контекст. Цепочка вызовов функций между ними должна
	распростронять Context, опционально заменяя его контекстом, полученным используя
	методы WithCancel, WithDeadline, WithTimeout или WithValue. Тогда при отмене
	контекста, полученные из него контексты так же будут отменены.

Снова из [Go блога: Context](https://blog.golang.org/context)

	В Google, мы требуем, что бы Go программисты добавляли параметр Context первым
	аргументом к каждой функции в пути вызовов между входящим и исходящим
	запросами. Это позволяет Go коду, созданному множеством различных групп
	нормально взаимодействовать. Это обеспечивает простой контроль времени ожидания 
	и отмены и гарантирует что критические значения, такие как учетные данные безопасности, корректно проходят через Go программы.

Задумайтесь на минутку о последствиях обладания каждой функции контекстом и удобстве этого.

Чувствуете неловкость? Хорошо. Давайте попытаемся следовать этому подходу, добавив `context` к нашему хранилищу, сделав ответственным его. Таким образом хранилище так же может передать контекст своим потомкам, которые так же будут ответственными за свою остановку.

## Сначал напишем тесты

Мы дложны будем изменить наши существующие тесты, поскольку их обязанности меняются. Единственной, за что сейчас отвечает наш обработчик - это проверка того, что он отправляет контекст в нижележащее хранилище и обрабатывает ошибку, которая будет поступать из хранилища при его отмене.

Давайте обновим `Store` для демонстрации новых возможностей.

```go
type Store interface {
	Fetch(ctx context.Context) (string, error)
}
```

Пока удалите код обработчика

```go
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
```

Обновите `SpyStore`

```go
func (s *SpyStore) Fetch(ctx context.Context) (string, error) {
	data := make(chan string, 1)

	go func() {
		var result string
		for _, c := range s.response {
			select {
			case <-ctx.Done():
				s.t.Log("spy store get cancelled")
				return
			default:
				time.Sleep(10 * time.Millisecond)
				result += string(c)
			}
		}
		data <- result
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-data:
		return res, nil
	}
}
```

Теперь наш наблюдатель действует так же, как реальный метод, работающий с `context`.

Мы симулируем медленный процесс путем медленной сборки результата строки символ за символом в горутине. Когда горутина завершится, она запишет результат в канал `data`. Так же горутина слушает канал ctx.Done, и остановит работу, если сигнал в канал поступит.

В конце кода, используется другой `select` для ожидания завершения работы, или для отмены.

Это похоже на наш предыдущий подход, когда мы использовали примитивы параллелизма Go, что бы два асинхронных процесса конкурировали друг с другом, что бы определить, что мы возвращаем.

Вы будуте использовать аналогичный подход при написании собственных функций и методов, которые принимают контекст, поэтому убедитель, что понимаете, что происходит.

В завершении обновим наши тесты. Закоментируйте тест отмены, что бы мы могли исправить тест успешного прохождения

## Запустите тест

    --- FAIL: TestHandler (0.00s)
        --- FAIL: TestHandler/returns_data_from_store (0.00s)
            my_test.go:65: got "", want "hello, world"

## Напишите минимальный код для его прохождения

```go
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := store.Fetch(r.Context())
		fmt.Fprint(w, data)
	}
}
```
Наш успешный код должне быть... успешным. Теперь мы можем исправить другие тесты.

## Сначала напишем тесты

Нам нужно проверить, что в случае ошибки мы не пишем никакого ответа. К сожалению, `http.ResponseRecorder` не может это выяснить, поэтому нам придется сыграть роль нашего собственного шпиоша, что бы проверить это

```go
type SpyResponseWriter struct {
	written bool
}

func (s *SpyResponseWriter) Header() http.Header {
	s.written = true
	return nil
}

func (s *SpyResponseWriter) Write([]byte) (int, error) {
	s.written = true
	return 0, errors.New("not implemented")
}

func (s *SpyResponseWriter) WriteHeader(staticCode int) {
	s.written = true
}
```

Наш `SpyResponseWriter` реализует `http.ResponseWriter`, так что мы можем его использовать в тестах

```go
	t.Run("tells store to cancel work if request is cancelled", func(t *testing.T) {
		store := &SpyStore{response: data, t: t}
		srv := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		cancellingCtx, cancel := context.WithCancel(request.Context())
		time.AfterFunc(5*time.Millisecond, cancel)
		request = request.WithContext(cancellingCtx)

		response := &SpyResponseWriter{}

		srv.ServeHTTP(response, request)

		if response.written {
			t.Error("a response should not have been written")
		}
```

## Запустите тест

    --- FAIL: TestHandler (0.13s)
    --- FAIL: TestHandler/tells_store_to_cancel_work_if_request_is_cancelled (0.01s)
        my_test.go:101: a response should not have been written

## Напишите минимальный код для его прохождения
```go
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := store.Fetch(r.Context())
		if err != nil {
			return
		}
		fmt.Fprint(w, data)
	}
}
```

Теперь мы можем видеть, что код сервера упрощен, поскольку он больше не несет явной ответственности за отмену

## Завершение 

### Что мы рассмотрели

 - Как тестировать HTTP обработчик, для которого запрос был отменен клиентом
 - Как использовать контекст для управления отменами
 - Как написать функцию, принимающую контекст, и использующий его для прерывания своей работы, используя горутины, `select` и каналы
 - Следовали руководству Google, говорящему, как управлять отменами путем распростронения контакста запроса через стек вызовов.
 - Как реализовать собственный http.ResponseWriter при необходимости

### Как насчет context.Value?

[Michal Strba](https://faiface.github.io/post/context-should-go-away-go2/) и я имеем похожее мнение 

    Если вы используете ctx.Value в моей (не существующей) компании, вы уволены

Некоторые программисы выступают за передачу значений через `context`, поскольку это *удобно*

Удобство часто является причиной плохого кода.

Проблема с `context.Values` заключается в том, что это просто нетипизированный map, и вам необходимо проверять, содержится ли необходимое в нем значение. Вы должны создавать связь ключей от модуля к модулю, и если кто то изменит что то, то начну происходить сбои.

В кратце, **Если функции необходимо некоторое значение, укажите его как типизированный параметр вместо того, что бы пытатья получить его из `Context.Value`**

**Но...**
С другой стороны, может быть полезно включать информацию, которая ортоганальна запросу в контексте, такую как идентификатор троссировки. Потенциально, эта информация не будет нужна каждой функции в вашем стеке вызовов и делает сигнатуры функций очень грязными.

[Джек Линдамуд говорт: **`Context.Value` должен уведомлять, а не управлять**](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)

    Содержимое context.Value для поддержки, не пользователей. Никогда не требуется ввод данных для документированных, или ожидаемых результатов.

## Дополнительный материал
 - [Context не должен быть в Go 2](https://faiface.github.io/post/context-should-go-away-go2/)
 - [Go Blog](https://blog.golang.org/context)