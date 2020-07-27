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