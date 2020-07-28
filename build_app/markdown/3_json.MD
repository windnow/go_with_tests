JSON, маршрутизация и встраивание
---------------------------------

В предыдущей главе мы создали web сервер для хранения количества выигранных играком игр.

У владельца появилось новое требование; иметь новый эндпойнт, с именем `/league` который возвращает список всех сохраненных игроков. Он хотел бы что бы возвращалось JSON значение.

Продолжим с того места, где остановились в прошлой главе.

## Сначала напишем тесты

Мы расширим имеющийся пакет, так как у нас есть несколько готовых полезных тестовых функций и тестовая реализация `PlayerStore`.

```go
func TestLeague(t *testing.T) {
	store := StubPlayerStore{}
	server := gameserver.NewServer(&store)

	t.Run("it return 200 on /league", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusOK)
	})
}
```

Прежде чем беспокоится получении результата и JSON, попытаемся придерживаться минимальных изменений в итерациях к нашей цели. Простейшее начало это проверить, можем ли мы обратившись к `/league` получить код возврата `OK`

## Попыткайтесь запустить тест

    windnow:
    т.к. по видимому эта часть руководства отличатеся от того, что мы реализовали в 
    прошлой главе, то у меня провал теста выдает "несоответствие кодов 404 и 200, 
    тогда как у автора код вызывает панику времени выполнения по причине того, что 
    используется слайс, тогда как в прошлой главе игрока мы получали через strings.
    TrimPrefix. 
    Далее вывод теста автора

```
=== RUN   TestLeague/it_returns_200_on_/league
panic: runtime error: slice bounds out of range [recovered]
    panic: runtime error: slice bounds out of range

goroutine 6 [running]:
testing.tRunner.func1(0xc42010c3c0)
    /usr/local/Cellar/go/1.10/libexec/src/testing/testing.go:742 +0x29d
panic(0x1274d60, 0x1438240)
    /usr/local/Cellar/go/1.10/libexec/src/runtime/panic.go:505 +0x229
github.com/quii/learn-go-with-tests/json-and-io/v2.(*PlayerServer).ServeHTTP(0xc420048d30, 0x12fc1c0, 0xc420010940, 0xc420116000)
    /Users/quii/go/src/github.com/quii/learn-go-with-tests/json-and-io/v2/server.go:20 +0xec
```
    windnow:
    Соответственно, ошибка out of range происходит, поскольку у автора имя игрока 
    определяется с помощью
     
    player := r.URL.Path[len("/players/"):]

    соответственно, для пустой строки будет вызываться паника, т.к. индекс, равный
    длине строки "/plyaers/" всего будет за пределами среза (out of range)

    Избавится от паники можно, либо использовав strings.TrimPrefix. Однако, для обработки разных роутов показано далее автором.

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	router := http.NewServeMux()
	router.Handle("/league", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	router.Handle("/players/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		player := strings.TrimPrefix(r.URL.Path, "/players/")

		switch r.Method {
		case http.MethodPost:
			p.processWin(w, player)
		case http.MethodGet:
			p.showScore(w, player)
		}
	}))

	router.ServeHTTP(w, r)
}
```

 - В начале обработки запроса мы создаем роутер, которому говорим, что для пути `x` использовать обработчик `y`.
 - Поэтому, для нашего пового эндпойнта мы используем `http.HandlerFunc` и *анонимную функцию*, для `w.WriteHeader(http.StatusOK)`, при запросе к `/league`, что бы наш тест прошел
 - Для эндпойнта /players/ мы переместили код в другой адаптер `http.HandlerFunc`.
 - В завершении мы обрабатываем полученный запрос с помощью вызова метода `ServeHTTP` нашего нового роутера (обрат внимание, что `ServeMYX` это так же `http.Handler`)

Теперь тест должен пройти

## Рефакторинг

`ServeHTTP` выглядит несколько великовато. Мы можем разделить путем выноса обработчиков в отдельные функции

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	router := http.NewServeMux()

	router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(p.playersHandler))

	router.ServeHTTP(w, r)
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (p *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	switch r.Method {
	case http.MethodPost:
		p.processWin(w, player)
	case http.MethodGet:
		p.showScore(w, player)
	}
}
```

Весьма странно (и не эффективно) настраивать роутер при поступлении запроса, а затем вызывать его. В идеале, мы хтим иметь функцию NwPlayerServer, которая примет наши зависимости и выполнит единовременное создание и настройку маршрутизатора. Затем каждый запрос сможет использовать этот экземплар маршрутизатора.

```go
// PlayerServer ...
type PlayerServer struct {
	store  PlayerStore
	router *http.ServeMux
}

// NewServer ...
func NewServer(store PlayerStore) *PlayerServer {
	p := &PlayerServer{
		store:  store,
		router: http.NewServeMux(),
	}

	p.router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	p.router.Handle("/players/", http.HandlerFunc(p.playersHandler))

	return p
}

func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
```
 - Теперь `PlayerServer` должен хранить маршрутизатор
 - Мы переместили создание маршрутизатора из `ServeHTTP` в `NewServer`, так что теперь это делается однажды, а не при каждом запросе.
 - Необходимо обновить весь код и тесты, где использовалось `PlayerServer{&store}` на `NewServer(&store)`

## Одно завершающее изменение

Попробуйте изменить код на следующий

```go
func NewServer(store PlayerStore) *PlayerServer {

	p := new(PlayerServer)
	p.store = store

	router := http.NewServeMux()
	router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(p.playersHandler))

	p.Handler = router

	return p
}
```

и удалите `func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request)`. Она нам больше не нужна!

## Встраивание

Мы изменили второе свойство `PlayerServer`, удалив именованное свойство `router http.ServeMUX` и заменили его на `http.Handler`; это называется *встраивание*

    Go не предоставляет обычную, основанное на типах, понятие 
	подклассов. Но у него есть возможность «заимствовать» части
	реализации внедряя типы в структуру или интерфейс.

[Effective Go - Embedding](https://golang.org/doc/effective_go.html#embedding)

Для нашего `PlayerServer` это значит что он получит все методы, имеющиеся в `http.Handler` имеющий лишь `ServeHTTP`.

Мы назначаем на него маршрутизатор, создаваемый в `NewServer`, поскольку `http.ServeMux` имеет метод `ServeHTTP`.

Это позволяет нам избавится от нашего метода `ServeHTTP`, так как он устанавливается через встроенный тип.

Встраивание это очень интересная возможност языка. Вы можете использовать его с интерфейсами для комбинирования их в новых интерфесах

```go
type Animal interface {
	Eater
	Sleeper
}
```

Так же встаривание можно использовать с конкретными типами, а не только с интефейсами, что позволит иметь доступ ко всем публичным методам и свойствам этого типа.

## Какие недостатки

Нужно быть осторожным со встраиванием, поскольку будут открыты все публичные методы и свойства встраиваемого типа. В нашем случае все нормально, поскольку встраивается лишь интерфейс, который нам нужно открыть (`http.Handler`) 

Если бы мы поленились, и встроили `http.ServeMux` вмето этого (конкретный тип), это тоже работало бы, однако пользователи `PlayerServer` могли бы добавить новые маршруты к нашему серверу, т.к. `Handle(paht, handler)` является публичным

**При встраивании типов задумайтесь, как это повлияет на на публичный API**

Злоупотребление встраиванием - очень распростроненая ошибка, поскольку это добавляет мусора к API и раскрывает внутренности типа.

Теперь, после реструктуризации нашего приложения мы легко можем добавлять новые маршруты и начнем с `/league`. Нужно что бы он возвращал некоторую полезную информацию.

Мы должны возвращать некоторый JSON, выглядящий как то так:

```json
[
	{
		"Name": Bill,
		"Wins":10
	},
	{
		"Name":"Alice",
		"Wins":15
	}
]
```


## Сначала напишем тесты

Начнем с попытки получить из ответа что то значимое

```go
func TestLeague(t *testing.T) {
	store := StubPlayerStore{}
	server := gameserver.NewServer(&store)

	t.Run("it return 200 on /league", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got []gameserver.Player

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("Unable to parse response from server %q into slice of Player, `%v`", response.Body, err)
		}

		assertStatusCode(t, response.Code, http.StatusOK)
	})

}
```

## Почему не тестируем строку JSON?

Вы могли бы возразить, что более простым начальным тестом было бы проверить на соответствие тела ответа некоторой строке JSON

По моему опыту, тесты для строки JSON имеют следующие проблемы:

 - *Хрупкость*: Если будет изменена модель данных, тест будет провален
 - *Сложность отладки*: Может быть сложно понять, в чем реальная проблема при сравнивании двух строк JSON
 - *Не верное стремление*: Хотя вывод и должен быть в JSON, на самом деле важно то, какие данные это данные, а не то, как они закодированы
 - *Повторный тест стандартной библиотеки*: Нет необходимости проверять, какой JSON возвращает стандартная библиотека. Она уже протестирована. Не тестируйте код других людей.

вместо этого нам нужно разобрать JSON в структуру данных, которые нам нужны для тестирования.

## Моделирование данных

Учитывая структуру JSON, похоже что нам нужен массив `Player` с некоторыми полями полями, поэтому создадим новый тип для получения этого.

```go
type Player struct {
	Name string
	Wins int
}
```

## Разбор JSON

```go
	var got []gameserver.Player
	err := json.NewDecoder(response.Body).Decode(&got)
```

Для разбора JSON в нашу модель данных мы создаем `Decoder` из пакета `encoding/json`, вызвав затем метод `Decode`. Декодуру нужно передать `io.Reader`, откуда будут читаться данные. В нашем случае это тело ответа `Body`

`Decode` принимает объекта, в который будет предприята попытка распарсить. Поэтому мы создаем пустой срез `Player` строкой выше.

Парсинг JSON с помощью `Decode` может вызвать ошибку, в случае которой нет смысла продолжать тестирование, поэтому, в случае ошибки, мы вызываем `t.Fatalf`. Обратите вниимание что тело ответа выводится, что бы проводящий тест мог понять, какая строка не может быть проанализирована.

## Попытайтесь запустить тест

```
--- FAIL: TestLeague (0.00s)
    --- FAIL: TestLeague/it_return_200_on_/league (0.00s)
        server_test.go:123: Unable to parse response from server "" into slice of Player, `EOF`
```

Наш эндпойнт пока не возвращает тела ответа, поэтому оно не может быть распаршено в JSON.

## Напишем минимум кода для прохождения теста

```go
func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	leagueTable := []Player{
		{"Chris", 20},
	}
	json.NewEncoder(w).Encode(leagueTable)

	w.WriteHeader(http.StatusOK)
}
```

Теперь тест пройдет

## Кодирование и декодирование

Обратите вниманиена прекрасную симметрию в стандартной библиотеке

 - Для создания `Encoder`'а необходим `io.Writer`, который реализует `http.ResponseWriter`
 - Для создания `Decodr`'а необходим `io.Reader`, которые реализует поле `Body` нашего наблюдателя.

В этой книге мы уже использовали `io.Writer`, что еще раз демонстрирует его распространенность в стандартной библиотеке и то, как многие библиотеки легко работают с ним.

## Рефакторинг

Было бы неплохо отделить получение `leagueTable` от нашего обработчика, поскольку известно, что вскоре мы не будет получать жестко закодированные значения.

```go
func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.getLeagueTable())

	w.WriteHeader(http.StatusOK)
}

func (p *PlayerServer) getLeagueTable() []Player {
	return []Player{
		{"Chris", 20},
	}
}
```

Далее мы хотим расширить нак тест т.о., что бы мы могли контролировать, какие данные должны быть получены.

## Сначала напишем тест

Мы можем обновить тест, что бы проверить, содержит ли таблицы лиги некоторого пользователя, которого мы держим в харнилище.

Обновите `StubPlayerStore`, что бы позволить хранить в нем лигу, являющиейся срезом объектов `Player`. 

```go
type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string
	league   []gameserver.Player
}
```

Далее, одновите текущий тест, поместив некоторых игроков и свойство лиги тестового хранилища, и проверьте, что они получены с сервера

```go
func TestLeague(t *testing.T) {

	t.Run("it returns thye league table as JSON", func(t *testing.T) {
		wantedLeague := []gs.Player{
			{"Cleo", 32},
			{"Chris", 20},
			{"Tiest", 14},
		}
		store := StubPlayerStore{nil, nil, wantedLeague}
		server := gs.NewServer(&store)
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got []gs.Player

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("Unable to parse response from server %q into slice of Player, %v", response.Body, err)
		}

		assertStatusCode(t, response.Code, http.StatusOK)

		if !reflect.DeepEqual(got, wantedLeague) {
			t.Errorf("got %v, want %v", got, wantedLeague)
		}
	})

}
```


## Запустите тест

```
./server_test.go:33:3: too few values in struct initializer
./server_test.go:70:3: too few values in struct initializer
```

## Напишите минимальный код для запуска тестов и получения ошибки тестирования

Нужно обновить другие тесты, поскольку появилось новое поле в `StubPlayerStore`; установите его в значение nil для других тестов

При повторном запуске теста вы должны получить 

```
--- FAIL: TestLeague (0.00s)
    --- FAIL: TestLeague/it_returns_thye_league_table_as_JSON (0.00s)
        server_test.go:138: got [{Chris 20}], want [{Cleo 32} {Chris 20} {Tiest 14}
```

## Напишем минимум кода для прохождения теста

Мы значем, что данные находятся в `StubPlayerStore`, являющийся реализацией абстракции интерфейса `PlayerStore`. Нам нужно обновить его, что бы любая реализация `PlayerStore` могла предоставить данные лиги

```go
type PlayerStore interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
	GetLeague() []Player
}
```

Теперь мы можем обновить наш обработчик, что бы при вызове он возвращал захардкоженный список. Удалите метод `getLeagueTable()` затем обновите `leagueHandler`, что бы он вызывал `GetLeague()`.

```go
func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.store.GetLeague())

	w.WriteHeader(http.StatusOK)
}
```

## Попытайтесь запустить тест.

Компилятор выдаст ошибку, поскольку для `InMemoryPlayerStore` и `StubPlayerStore` не реализован новый метод, добавленный в интерфейс.

Для `StubPlayerStore` исправить легко. Просто верните поле `league` добавленное ранее.

```go
func (s *StubPlayerStore) GetLeague() []gs.Player {
	return s.league
}
```

Вот так выглядит реализация `InMemoryPlayerStore`

```go
type InMemoryPlayerStore struct {
	store map[string]int
}
```

Хотя было бы не сложно «дложным образом» реализовать `GetLeague` перебирая карту, помните, что мы пытаемся писать минимальный объем кода, что бы тесты работали.

```go
func (i *InMemoryPlayerStore) GetLeague() []gs.Player {
	return nil
}
```

На самом деле это говорит нам, что позже нужно будет реализовать это, однако пока давайте оставим это

Попытайтесь запустить тесты. Компиляция должна пройти успешно, так же как и тесты

## Рефакторинг

Тествоый код не очень хорошо выражает намерение, и в нем есть много шаблонного кода, который можно переработать

```go
func TestLeague(t *testing.T) {

	t.Run("it returns thye league table as JSON", func(t *testing.T) {
		wantedLeague := []gs.Player{
			{"Cleo", 32},
			{"Chris", 20},
			{"Tiest", 14},
		}
		store := StubPlayerStore{nil, nil, wantedLeague}
		server := gs.NewServer(&store)
		request := newLeagueRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getLeagueFromResponse(t, response.Body)
		assertStatusCode(t, response.Code, http.StatusOK)
		assertLeague(t, got, wantedLeague)
	})

}

func assertLeague(t *testing.T, got, want []gs.Player) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func newLeagueRequest() *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/league", nil)
	return request
}

func getLeagueFromResponse(t *testing.T, body io.Reader) (league []gs.Player) {
	t.Helper()

	err := json.NewDecoder(body).Decode(&league)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, %v", body, err)
	}

	return league

}
```

Последнее, что мы должны сделать, что бы наш сервер заработал - это убедиться, что возвращается заголовок с `content-type`, по которому компьютер мог бы понять что возвращается `JSON`

## Сначала напишем тесты

Добавьте это сравнение в существующий тест
```go
		if response.Result().Header.Get("content-type") != "application/json" {
			t.Errorf("response did not have content-type of application/json, got %v", response.Result().Header)
		}
```

## Запустите тест

```
    --- FAIL: TestLeague/it_returns_thye_league_table_as_JSON (0.00s)
        server_test.go:137: response did not have content-type of application/json, got map[Content-Type:[text/plain; charset=utf-8]]
```

## Напишите достаточное количество кода для его прохождения

Обновите `leagueHandler`

```go
func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(p.store.GetLeague())
}
```

Теперь тест должен пройти

## Рефакторинг

Добавим хелпер для `assertContentType`:

```go
const jsonContentType = "application/json"

func assertContentType(t *testing.T, response httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of application/json, got %v", want)
	}

}
```

затем задействуте в тесте

```go
		assertContentType(t, response, jsonContentType)
```

Теперь, кодга мы уже разобрались с `PlayerServer`, мы можем переключиться на `InMemoryPlayerStore`, потому что если сейчас мы покажем код заказчику, то `/league` не будет работать

Самый быстрый способ проверки это добавить в имеющийся интеграционный тест. Мы можем указать новый эндпойнт, и проверить что получаем корректный ответ с `/league`.

## Сначала напишем тесты

Можно использовать `t.Run` что бы разделить этот тест, а так же для повторного использования хелперов - снова демонстрируя важность рефакторинга тестов

```go
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	store := inmemstore.NewInMemoryStore()
	server := gs.NewServer(store)
	player := "Pepper"

	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))

	t.Run("get score", func(t *testing.T) {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, newGetScoreRequest(player))
		assertStatusCode(t, response.Code, http.StatusOK)

		assertResponseBody(t, response.Body.String(), "3")
	})

	t.Run("get league", func(t *testing.T) {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, newLeagueRequest())
		assertStatusCode(t, response.Code, http.StatusOK)

		got := getLeagueFromResponse(t, response.Body)
		want := []gs.Player{
			{player, 3},
		}

		assertLeague(t, got, want)
	})
}
```

## Запустите тест

```
   --- FAIL: TestRecordingWinsAndRetrievingThem/get_league (0.00s)
        server_test.go:131: got [], want [{Pepper 3}]
```

## Напишем код для его прохождения

`InMemoryPlayerStore` возвращает `nil` при вызове `GetLeague()`, поэтому нам нужно исправить его

```go
func (i *InMemoryPlayerStore) GetLeague() []gs.Player {
	var league []gs.Player

	for name, wins := range i.store {
		league = append(league, gs.Player{name, wins})
	}
	return league
}
```

Все, что нам нужно сделать, это пройтис по карте и сконвертировать каждую пару ключ/значение в `Player`

Теперь тест должен пройти

-------------------------------------------------------

## Завершение

Мы продолжили итеративно развивать нашу программу и использованием TDD, добавляя в нее возможность поддержки новых эндпойнтов в маршрутизаторе и теперь она может возвращать JSON для наших клиентов. В следующей главе мы рассмотрим хранение данных и сортировку лиги.

Что мы рассмотрели:

 - **Роутинг**. Стандартная библиотека предлагает простой в использовании тип для маршрутизации. Его полностью охватывает интерфейс `http.Handler`, в который назначаются маршруты для обработки и маршрутизации. Он не имеет некоторых функций, которые можно было бы ожидать, таких как переменные пути (н.р. /users/{id}). Можно с легкостью реализовать этот функционал самостоятельно, а так же можно рассмотреть возможность использования других библиотек маршрутизации, если возникнут сложности в разработке своей. Большинство популярных придерживается стандартного подхода, и так же реализуют `http.Handler`
 - **Встраивание типов**. Мы частично коснулись этой техники, но [вы можете самостоятельно у знать о не из Effective Go](https://golang.org/doc/effective_go.html#embedding). Если есть некоторый функционал, который хотелось бы заимствовать, встраивание может оказаться чрезвычайно полезной, но не следует забывать о публичном API; раскрывайте только то, что подходит.

 - **Сериализация/десериализация JSON**. Стандартная библиотека делает обыденной эту задачу. Кроме того, она настраиваемая и можно конфигурировать эти преобразования при необходимости.