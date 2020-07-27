## HTTP server

Вас попросили создать web сервер, который позволяет пользователям отслеживать количество выигранных пользователем игр.

 - `GET /players/{name}` должно возвращать общее число побед
 - `POST /players/{name}` должно записывать победу для данного имени, инкрементируя для каждого последующего `POST` запроса 

Мы будем следовать TDD подходу, получая работающий вариант как можно скорее, внося затем небольшие итеративные измененя, пока не получим решение. Следуя этому подходу мы:

 - Поддерживать минимальное пространсво задачи в любой момент времени
 - Не создавайте проблемы
 - Если мы где то запутаемся, откат не приведет к потере работы

## Красный, зеленый, рефакторинг

В этой книге мы акцентируемся на процессе TDD написания тестов и отслеживания ошибок (красный), написании минимального кода для его работы (зеленый) и рефакторинге

Правило написания минимального количесвта кода важно с точки зрения безопасности, которую вам дает TDD. Вы должны стремится как можно скорее избавиться от "красного" состояния.

Кент Бек описывает это как:

    Заставляйть тесты работать как можно скорее, допуская любые огрехи в процессе

## Что, если мы не будем так делать?

Чем больше изменений вы сделаете, пока находитесь в красном состоянии, тем больше вероятность того, что вы добавите больше проблем, не охваченных тестами.

Идея заключается в итеративном написании полезного кода малыми шагами, управляемого тестами, т.о. вы не создадите проблемы, с которыми будете работать часами.

## Курица и яйцо

Каким образом инкрементно разрабатывать? Мы не можем отправить `GET` запрос для игрока без имеющегося сохраненного значения, и трудно понять, сработал ли `POST` запрос без существующей конечной точки `GET`

Вот тут и появляются моки.

 - `GET` запросу понадобится `PlayerStore` для получения счета игрока. Это должнен быть интерфейс, и при тестировании мы може создать небольшую реализацию для тестирования нашего кода, без необходимости иметь реализацию кода хранилища.
 - для `POST` запроса мы можем наблюдать за его вызовами для проверки корректности хранения игроков
 - Для получения некоторого работающего приложения, мы можем реализовать простую реализацию в памяти и позднее можем создать реализацию любого выбранного механизма хранилища


## Для начала напишем тесты

Мы можем написать тест и заставить его пройти захардкодив для начала нужные значения. Кент Бек называет это "подделкой". Как только у нас будут работающие тесты, мы сможем написать больше тестов, которые помогут нам избавиться от констант

Сделав этот очень маленький шаг, мы можем сделать важный старт для правильной работы всей структуры проекта, не беспокоясь о логике нашего приложения.

Для создания web сервера на Go обычно вызывается функция [`ListenAndServe`](https://golang.org/pkg/net/http/#ListenAndServe)

```go
func ListenAndServe(addr string, handler Handler) error
```

Это запустить сервер, слушающий порт и создающий горутину для каждого запроса и запускающий его вместе с обработчиком [`Handler`](https://golang.org/pkg/net/http/#Handler).

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Тип реализует интерфейс Handler, предоставляющий метод `ServerHTTP`, принимающий два аргумента, в первый мы *пишем наш ответ* и второй это HTTP запрос, отправленный серверу.

Давайте напишем тест для функции `PlayerServer` принимающий эти два аргумента. Запрос будет состоять в том, чтобы получить счет игрока, который мы ожидаем равным "20" 

```go
func TestGETPlayers(t *testing.T) {
	t.Run("returns Pepper's score", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/players/Pepper", nil)
		response := httptest.NewRecorder()

		PlayerServer(response, request)

		got := response.Body.String()
		want := "20"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
```

Что бы проверить наш сервер, нам потребуется запрос на отправку, и нам нужно следить за тем, что наш обработчик пишет в `ResponseWriter`

 - Мы используем `http.NewRequest` для создания запроса. Первый аргумент это метод запроса, а второй это путь запроса. Знкачение `nil` посылаемое вместо тела запроса, которое нам не нужно в этом тесте.
 - `net/http/httptest` имеет готового шпиона для нас, называемый `ResponseRecorder`, так что мы можем использовать его. Он имеет множество полезных методов, чтобы проверить, что было записано как ответ.


## Попытайтесь запустить тест

```bash
internal/server/server_test.go:14:3: undefined: PlayerServer
```

## Напишите минимальное количество кода для запуска теста и проверьте его результат

```go
func PlayerServer() {}
```

и испытайте тест снова

```bash
internal/server/server_test.go:16:22: too many arguments in call to server.PlayerServer
        have (*httptest.ResponseRecorder, *http.Request)
        want ()
```

Добавьте аргументы к нашей функции

```go
func PlayerServer(w http.ResponseWriter, r *http.Request) {}
```

Теперь тест скомпелируется и тест провалится

```bash
--- FAIL: TestGETPlayers (0.00s)
    --- FAIL: TestGETPlayers/returns_Pepper's_score (0.00s)
        server_test.go:22: got "", want "20"
```		

## Напишите минимальный код для его прохождения

В главе DI мы коснулись HTTP сервера с функцией `Greet`. Мы узнали что `ResponseWriter` из net/http так же реализует `Writer`, так что мы можем использовать `fmt.Fprint` для отправки строки как HTTP ответ

```go
func PlayerServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "20")
}
```

Теперь тест должен пройти

## Завершите конструкцию

Мы хотим подключить это в приложение. Это важно, поскольку

 - У нас будет рельно работающее приложение. Мы не хотим писать тесты ради тестов, хорошо видеть код в действии
 - По мере рефакторинга кода структура программы будет меняться. Мы хотим что бы это отображалось в нашей программе, как часть инкрементального подхода

Создайте новый файл для нашего приложения и поместите в него этот код

```go
package main

import (
	"log"
	"net/http"

	"github.com/windnow/edusrv/internal/server"
)

func main() {
	handler := http.HandlerFunc(server.PlayerServer)
	if err := http.ListenAndServe(":5000", handler); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
```

До сих пор весь код нашего приложения был в одном файле, однако, это не лучшая практика для больших приложений, где вы захотите разделить код на разные файлы

Для запуска, выполните `go build`, который возьмет все `.go` файлы в дирректории и соберет программу. Вы можете ее запустить командой `./myprogram`

## http.HandlerFunc

Ранее мы рассмотрели, что необходимо реализовать интерфейс `Handler` для создания `server`а
для создания `server`а. *Обычно* мы делаем это создав `struct` реализующий метод интерфейса `ServeHTTP`. Однако, вариант хранения со структурами предназначен для хранения данных, но в *настоящее* время у нас нет состояния, поэтому его не очень удобно создавать.

[`HandlerFunc`](https://golang.org/pkg/net/http/#HandlerFunc) позволяет нам избежать этого

	Тип HandlerFunc это адаптер, позволяюий использовать обычные функции в качестве
	обработчиков HTTP. Если f является функцией с соответсвующей сигнатурой
	HandlerFunc(f)	это обработчик, вызывающий f.

```go
type HandlerFunc func(resonseWriter, *Request)
```

Из документации видно, что тип `HandlerFunc` уже реализует метод `ServeHTTP`. Передав в него нашу функцию `PlayerServer`, мы реализовали требуемый обработчик

## http.ListenAndServe(":5000"...)

`ListenAndServe` принимает порт прослушивания для обработчикаю. Если порт будет занят, то будет возвращена ошибка, поэтому мы используем выражение `if` для её захвата и логирования проблемы для пользователя

Следующее, что мы сделаем - напишем еще один тест, что бы внести некоторые позитивные изменения, попытавшить уйти от жестко закодированных значений

## Сначал напишем тест

Добавим еще один подтест, который попытается получить очки для другого игрока, что испортит наш жестко закодированный подход

```go
	t.Run("return Floyd's score", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/players/Floyd", nil)
		response := httptest.NewRecorder()

		server.PlayerServer(response, request)

		got := response.Body.String()
		want := "10"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
```

Вы можете подумать

	Разумеется, нам нужно некоторая концепция хранилища, что бы контролировать, 
	какой игрок получает какой счет. Странно, что значения кажутся такими 
	произвольными в наших тестах

Помните, что мы просто стараемся делать шаги как можно меньше, так что теперь попытаемся избавится от констант

## Попытайтесь запустить тесты

```
=== RUN   TestGETPlayers/return_Floyd's_score
    TestGETPlayers/return_Floyd's_score: server_test.go:35: got "20", want "10"
--- FAIL: TestGETPlayers (0.00s)
    --- PASS: TestGETPlayers/returns_Pepper's_score (0.00s)
    --- FAIL: TestGETPlayers/return_Floyd's_score (0.00s)

```

## Напишем минимальный код для прохождения теста

```go
func PlayerServer(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	if player == "Pepper" {
		fmt.Fprint(w, "20")
		return
	}
	if player == "Floyd" {
		fmt.Fprint(w, "10")
		return
	}
}
```

Этот тест заставил нас обратить внимание на URL запроса и принять решение от его значения. Так что пока мы беспокоимся о хранилище и интерфейсе для игроков, следующий логический шаг, по видимому, касается *маршрутизации*.

Если бы начали с кода хранилища, количество изменений, которое нам бы пришлось сделать было бы очень большим по сравнению с роутингом. **Меньший шаг к нашей конечной цели обусловлен тестами**

## Рефакторинг

Мы можем упростить `PlayerServer` выделив получение оценки в отдельную функцию

```go
func PlayerServer(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	fmt.Fprint(w, GetPlayerScore(player))

}

// GetPlayerScore ...
func GetPlayerScore(name string) string {
	if name == "Pepper" {
		return "20"
	}
	if name == "Floyd" {
		return "10"
	}
	return ""
}
```

И мы можем улучшить часть кода в тестах по DRY, добавив несколько хелперов

```go
func TestGETPlayers(t *testing.T) {
	t.Run("returns Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.PlayerServer(response, request)

		assertResponseBody(t, response.Body.String(), "20")
	})
	t.Run("return Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.PlayerServer(response, request)

		assertResponseBody(t, response.Body.String(), "10")
	})
}

func newGetScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func assertResponseBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}
```

Однако мы еще не можем успокотся. Неправильно, что наш сервер знает результаты

Наш рефакторинг дал понять, что делать

Мы перенесли вычисление счета из тела основного обработчика в функцию `GetPlayerScore`. Это похоже на подходящее место для разделения проблем с помощью интерфейсов.

Давайте переместим функцию, что бы появился интерфейс

```go
type PlayerStore interface {
	GetPlayerScore(name string) int
}
```

Что бы наш `PlayerServer` мог использовать `PlayerStore`, ему потребуется ссылка на него, так что пришло время изменить архитектуру т.о. что бы наш `PlayerServer` стал структурой

```go
type PlayerServer struct {
	store PlayerStore
}
```

В завершении реализуем интерфейс `Handler` добавив метод к нашей новой структуре добавив существующий код обработчика

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")
	fmt.Fprint(w, p.store.GetPlayerScore(player))
}
```

Единственное изменение, которое мы сделали - вызов `store.GetPlayerScore` для получения оценки вместо локальной функции, которую теперь можно удалить

Вот полный код нашего сервера

```go
type PlayerStore interface {
	GetPlayerScore(name string) int
}

type PlayerServer struct {
	store PlayerStore
}

func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")
	fmt.Fprint(w, p.store.GetPlayerScore(player))
}
```

## Исправление проблем

Мы сделали много изменений, и знаем что наши тесты и приложение большене будут компилироваться. Но не переживайте и позвольте компилятору поработать над этим

```
cmd/gamelogger/main.go:11:29: type server.PlayerServer is not an expression
```

Нам нужно изменить наши тесты создав экземпляр `PlayerServer` вызвав затем метод `ServeHTTP`

```go
func TestGETPlayers(t *testing.T) {
	server := &gameserver.PlayerServer{}
	t.Run("returns Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "20")
	})
	t.Run("return Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "10")
	})
}
```

Обратите внимание - мы все еще не беспокоимся о реализации хранилища, а просто хотим что бы компиляция проходила как можно скорее.

У вас дложна выработатья привычка расставлять приоритеты сначала получив код, который компилируется, затем код, который проходит тесты

Добавляя больше функциональности (такой как часть хранилища), пока код не компилируется мы открываем себя для *больших* проблем компиляции.

Теперь `main.go` тоже не будет компилироваться по той же причине

```go
func main() {
	server := &gameserver.PlayerServer{}

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
```

Наконец все компилируется, однако тесты проваливаются

```
--- FAIL: TestGETPlayers (0.00s)
    --- FAIL: TestGETPlayers/returns_Pepper's_score (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
        panic: runtime error: invalid memory address or nil pointer dereference
```

Это потому, что мы не указали `PlayerStore` в наших тестах. Необходимо реализовать заглушку

```go
type StubPlayerStore struct {
	scores map[string]int
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	score := s.scores[name]
	return score
}
```

`map` является простым и быстрым способом реализовать заглушку типа ключ/значение для наших тестов. Теперь давайте создадим экземпляр этого хранилища и передадим в наш `PlayerServer`

```go
func TestGETPlayers(t *testing.T) {
	store := &StubPlayerStore{
		map[string]int{
			"Pepper": 20,
			"Floyd":  10,
		},
	}
	server := gameserver.NewServer(store)
	t.Run("returns Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "20")
	})
	t.Run("return Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "10")
	})
}
```

============================
т.к. сервер и тесты находятся в разных пакетах, а так же поле `store` структуры `PlayerServer` является приватным, то в пакете `gameserver` реализована функция, создающая сервер 
```go
func oewServer(store PlayerStore) *PlayerServer {
	return &PlayerServer{
		store: store,
	}
}
```
============================

Теперь наши тесты проходят, и выглядят лучше. Намерения, стоящие за нашим кодом, стало более понятным и благодара введения хранилища. Мы говорим читателю, что, поскольку эти данные хранятся в `PlayerStore`, то при использовании их с `PlayerServer` вы должны получить следующие ответы.

## Запуск приложения
Теперь, когда наши тесты работают, последнее, что нам следует сделать - это завершить рефакторинг для проверки работоспособности приложения. Программа должна запустится, однако вы получите плохой ответ, если попытаетесь обратится к серверу по адресу `http://localhost:5000/players/Pepper`.

Причина этого в том, что мы не передаем `PlayerStore`.

Необходимо создать его реализацию, однако пока это сложно поскольку мы не храним никаких значимых данных, поэтому пока их нужно будет захардкодить

Если вы снова соберете и запустите приложение перейдя затем по URL, то должны получить "123". Не здорово, однако пока мы не храним данные - это лучшее что мы можем сделать

У нас есть несколько вариантов того, что делать дальше:

 - Обработать сценарий, при котором игрок не существует
 - Обработать сценарий для метода `POST /players/{name}`
 - Было не хорошо, что основное приложение запустилось, но реально не работало. Обнаружить проблему мы должны были вручную.

Пока `POST` сценарий приближает нас к «успешному пути», поскольку мы уже в соответствующем контексте. До остального доберемся позже

## Сначала напишем тест

Добавте сценарий не найденного игрока в существующий тест

```go
	t.Run("returns 404 on missing players", func(t *testing.T) {
		request := newGetScoreRequest("Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusNotFound

		if got != want {
			t.Errorf("got status %d want %d", got, want)
		}
	})
```

## Запустите тест

```
    --- FAIL: TestGETPlayers/returns_404_on_missing_players (0.00s)
        server_test.go:56: got status 200 want 404
```

## Напишите минимальный код для его прохождения

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	w.WriteHeader(http.StatusNotFound)

	fmt.Fprint(w, p.store.GetPlayerScore(player))
}
```

Иногда я закатываю глаза, когда сторонники TDD говорят: «Убедитесь, что вы пишите минимальное количество кода, что бы тест прошел», поскольку это выглядит очень педантично

Но этот пример хорошо иллюстрирует сценарий. Я сделал минимум (понимая, что это не верно), т.е. написал `StatusNotFound` для **всех запросов**, однако все тесты проходят

**Реализуя минимум для прохождения тестов вы можете выделить пробемы в тестах**. В нашем случае мы не утверждаем, что всегда должны получать `StatusOK`, когда игрок имеется в хранилище

Обновите наши два теста для проверки статуса, и исправте код

Вот обновленный тест

```go
func TestGETPlayers(t *testing.T) {
	store := &StubPlayerStore{
		map[string]int{
			"Pepper": 20,
			"Floyd":  10,
		},
	}
	server := gameserver.NewServer(store)
	t.Run("returns Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "20")
		assertStatusCode(t, response.Code, http.StatusOK)
	})
	t.Run("return Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "10")
		assertStatusCode(t, response.Code, http.StatusOK)
	})

	t.Run("returns 404 on missing players", func(t *testing.T) {
		request := newGetScoreRequest("Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusNotFound)

	})
}

func newGetScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func assertResponseBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}

func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not correct status code. got %d, want %d", got, want)
	}
}
```

Мы проверяем статус в каждом из наших тестов, так что я сделал хелпер `asertStatusCode` для реализации этого

Теперь первые два теста провалятся, поскольку получают 404 вместо 200. Для исправления необходимо что бы `PlayerStore` возвращал «Не найдено» только в случае счета равного 0

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}
```

### Хранение счета

Теперь, когда мы можем получить счет из хранилища, имеет смысл реализовать хранение новых значений

## Сначала напишем тест

```go
func TestStoreWins(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{},
	}
	var server = gameserver.NewServer(&store)

	t.Run("it returns accepted on POST", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/players/Pepper", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusAccepted)
	})
}
```

Для начала проверим, получим ли верный код состояния, если попали в конкретный маршрут с помощью `POST`. Это позволяет исключить возможность принимать запросы другого типа и обрабатывать их иначе, чем `GET /players/{name}`. После того, как это сработает, мы можем начать проверять взаимодействие нашего обработчика с хранилищем

## Попытайтесь запустить тест
```
--- FAIL: TestStoreWins (0.00s)
    --- FAIL: TestStoreWins/it_returns_accepted_on_POST (0.00s)
        server_test.go:71: did not correct status code. got 404, want 202
```

## Напишите минимальный код для его прохождения
Запомните, что мы сознательно допускаем огрехи, поэтому оператор `if` основанный на методе запроса справится с задачей

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}
```

## Рефакторинг

Обработчик выглядит несколько запутанно. Для изоляции и более простого отслеживания давайте разделим функциональность в новые функции

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		p.processWin(w)
	case http.MethodGet:
		p.showScore(w, r)
	}

}

func (p *PlayerServer) showScore(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}
```

Теперь аспект маршрутизации стал немного понятнее, а так же наши следующие итерации при сохранении могут быть реализованы внутри `processWin`

Далее, мы хотем что бы при обращении к `POST /players/{name}` `PlayerStore` записывал результат

## Сначал напишем тесты

Мы можем реализовать это, расширив наш `StubPlayerStore` методом `RecordWin`, следя затем за его вызовами.

```go
type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	score := s.scores[name]
	return score
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.winCalls = append(s.winCalls, name)
}
```

Теперь расширим наш тест проверкой числа вызовов для начала

```go
func TestStoreWins(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{},
	}
	server := gameserver.NewServer(&store)

	t.Run("it records wins when POST", func(t *testing.T) {
		request := newPostWinRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusAccepted)

		if len(store.winCalls) != 1 {
			t.Errorf("got %d calls to RecordWin want %d", len(store.winCalls), 1)
		}
	})
}
func newPostWinRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return req
}
```

## Запустите тест

```
internal/gameserver/server_test.go:28:17: too few values in StubPlayerStore literal
internal/gameserver/server_test.go:66:17: too few values in StubPlayerStore literal
```

## Напишите минимальное количество кода для запуска теста и проверки сообщения ошибки теста

Нам нужно обновить наш код в месте создания `StubPlayerStore`, т.к. мы добавили новое поле

```go
store := StubPlayerStore{
		map[string]int{},
		nil,
	}
```

```
--- FAIL: TestStoreWins (0.00s)
    --- FAIL: TestStoreWins/it_records_wins_when_POST (0.00s)
        server_test.go:81: got 0 calls to RecordWin want 1
```

# Напишем минимум кода для прохождения теста

Поскольку мы определяем количество вызовов, а не конкретные значения, это делает нашу начальную итерацию несколько короче

Нам нужно обновить информацию о `PlayerStore` в `PlayerServer` изменив интерфейс т.о. что бы можно было вызвать `RecordWin`.

```go
type PlayerStore interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
}
```

Сделав это, основная программа больше не компелируется

```
cmd/gamelogger/main.go:17:33: cannot use &inMemoryPlayerStore literal (type *inMemoryPlayerStore) as type gameserver.PlayerStore in argument to gameserver.NewServer:
        *inMemoryPlayerStore does not implement gameserver.PlayerStore (missing RecordWin method)
```

Компилятор сообщает нам, что не так. Давайте обновим `InMemoryPlayerStore` реализовав этот метод.

```go
type inMemoryPlayerStore struct{}

func (i *inMemoryPlayerStore) RecordWin(name string) {}
```

Теперь, когда для `PlayerStore` имеется `RecordWin` мы можем вызвать его в нашем `PlayerServer`

```go
func (p *PlayerServer) processWin(w http.ResponseWriter) {
	p.store.RecordWin("Bob")
	w.WriteHeader(http.StatusAccepted)
}
```

Запустите тесты, и он должен пройти. Разумеется, `"Bob"` это не то, что мы хотим отправлять в `RecordWin`, так что давайте доработаем тест

## Сначала напишем тест

```go
	t.Run("it records wins on POST", func(t *testing.T) {
		player := "Pepper"

		request := newPostWinRequest(player)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusAccepted)

		if len(store.winCalls) != 1 {
			t.Errorf("got %d calls to RecordWin want %d", len(store.winCalls), 1)
		}

		if store.winCalls[0] != player {
			t.Errorf("did not store correct winner got %q want %q", store.winCalls[0], player)
		}
	})
```

Теперь, когда мы значем, что в нашем срезе `wincalls` есть один элемент. мы можем смело ссылаться на первый, и проверить его на равенство игроку

## Попытайтесь запустить тест

```
    --- FAIL: TestStoreWins/it_records_wins_on_POST (0.00s)
        server_test.go:87: did not store correct winner got "Bob" want "Pepper
```

## Напишите минимальный код для его прохождения

```go
func (p *PlayerServer) processWin(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")
	p.store.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}
```

Мы изменили `processWint` что бы он принимал `http.Request`, так что мы можем посмотреть на URL что бы извлечь имя игрока. Как только мы его получим, мы можем обратится к хранилищу с правильным значением, что бы пройти тест

## Рефакторинг

### DRY 
Мы можем немного подчистить код, т.к. извлекаем имя игрока в двух местах

```go
func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	player := strings.TrimPrefix(r.URL.Path, "/players/")

	switch r.Method {
	case http.MethodPost:
		p.processWin(w, player)
	case http.MethodGet:
		p.showScore(w, player)
	}

}

func (p *PlayerServer) showScore(w http.ResponseWriter, player string) {

	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, player string) {
	p.store.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}
```

Несмотря на то, что наши тесты проходят, на самом деле у нас нет работающей программы. Если вы попытаетесь запустить и изспользовать приложение по назначению, оно не будет работать, потому что у нас нет правильной реализации `PlayerStore`. Это хорошо, поскольку сосредоточившись на нашем обработчике, мы определили интерфейс, который нам нужен, а не пытались спроектировать его заранее

Мы могли бы начать писать некоторые тесты для нашего `InMemoryPlayerStore`, но только временно, пока мы не реализуем более надежный способ сохранения результатов игрока (т.е. базы данных)

Сейчас мы напишем интеграционный тест между нашими `PlayerServer` и `InMemoryPlayerStore`, что бы завершить работу. Это позволит нам быть уверенными в том, что наше приложение работает, без необходимости непосредственного тестирования `InMemoryPlayerStore`. Кроме того, когда мы приступим к реализации `PlayerStore` мы сможем протестировать эту реализацию с помощью того же интеграционного теста 

## Интеграционные тесты

Они могут быть полезны при преверке работы больших облостей вашей системы. Однако имейте в виду:

 - Их сложнее писать
 - Если они проваливаются, может быть сложно определить почему (обычно это ошибка в компоненте интеграционного теста), и поэтому их бывает труднее исправить.
 - Иногда они работают медленно (поскольку часто используют "реальные" компоненты, такие как базы данных)

По этим причинам рекомендуется изучить The Test Pyramid.

## Сначала напишем тест

В целях краткости я покажу вам конечный, отрефакторинговый вариант интеграционного теста 

```go
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	store := inmemstore.NewInMemoryStore()
	server := gameserver.NewServer(store)
	player := "Pepper"
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assertStatusCode(t, response.Code, http.StatusOK)

	assertResponseBody(t, response.Body.String(), "3")
}
```

## Запустите тест 

```
=== RUN   TestRecordingWinsAndRetrievingThem
    TestRecordingWinsAndRetrievingThem: server_test.go:105: response body is wrong, got "123" want "3"
--- FAIL: TestRecordingWinsAndRetrievingThem (0.00s)
```

## Напишите минимальный код для его прохождения

Здесь я пущусь в некоторую вольность и напишу больше кода, чем вы можете себе представить, без написания теста.

Это допускается, т.к. у нас есть тест, проверяющий работоспособность, хотя он и не отностися к конкретному модулю (`InMemoryPlayerStore`)

Если бы я где то в реализации застрял, то я бы вернул все обратно к провалу теста, а затем написал бы более конкретные тесты вокруг `InMemoryPlayerStore`, для помощи в поиске решения.

```go
type inMemoryPlayerStore struct {
	store map[string]int
}

func (i *inMemoryPlayerStore) RecordWin(name string) {
	i.store[name]++
}

func (i *inMemoryPlayerStore) GetPlayerScore(name string) int {
	return i.store[name]
}

func NewInMemoryStore() *inMemoryPlayerStore {
	return &inMemoryPlayerStore{
		map[string]int{},
	}
}
```


 - Нам нужно хранить данные, поэтому я добавил `map[string]int` в структуру `InMemoryPlayerStore`
 - Для удобства я добавил `NewInMemoryStore`, производящий инициализацию хранилища и обновил интеграционный тест для ее использования (`store := NewInMemoryStore()`)
 - Остальная часть кода просто оборачивает мапу

Интеграционный тест проходит, теперь нужно изменить `main` для использования `NewInMemoryStore()`

```go
package main

import (
	"log"
	"net/http"

	"github.com/windnow/edusrv/internal/gameserver"
	"github.com/windnow/edusrv/internal/inmemstore"
)

func main() {
	server := gameserver.NewServer(inmemstore.NewInMemoryStore())

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
```

Соберите его, и используйте `curl` или `httpie` для проверки

 - Запустите команду несколько раз, измените имя на любое
```sh
$ http POST http://localhost:5000/players/Pepper
```

 - Проверьте счет командой 
 ```sh
 http GET http://localhost:5000/players/Pepper
 ```
Круто! Вы сделали REST подобный сервис. Для его применения в будущем необходимо выбрать хранилище для хранения данных дольше, чем время выполнения программы

 - Выберите хранилище (Bold?, Mongo? Postgres? Файловая система?)
 - Создайте реализацию `PlayerStore` например, `PostgresPlayerStore`
 - Функциональность TDD позволяте быть уверенным в ее работоспособности
 - Добавьте его к интеграционному тесту. Проверьте, все ли в порядке.
 - В завершении, добавьте его в функцию `main`

## Рефакторинг
  
Мы почти закончили. Давайте предпримем некоторые усилия для предотвращения паралелизма, подобные этим: 

```
fatal error: concurrent map read and map write
```
Добавив мьютексы, мы обеспечиваем безопасность паралелизма, особенно, для счетчика в нашей функции `RecordWin`. Подробнее о мьютексах читайте в главе [синхронизации](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/sync)

## Завершение

### http.Handler

 - Реализуйте этот интерфейс для создания web серверов
 - Используйте `http.HandlerFunc`, чтобы превратить обычные функции в `http.Handler`
 - Используйте `httptest.NewRecorder` для передачи в качестве `ResponseWriter`, что позволит вас следить за ответами, которые отправляет обработчик
 - Используйте `http.NewRequest` для построения запросов, которые будут отправлены вашей системе

## Интерфейсы, имитация и DI

 - Позволяет итеративно строить систему небольшими порциями
 - Позволяют разрабатывать обработчики, которым необходимо хранилище без использования реального хранилища
 - TDD позволяет выявить необходимые интерфейсы

## Фиксируйте погрешности, затем рефакторьте (и затем добавляйте контроль версий)

 - Необходимо рассматривать неудачные компиляцию и тесты как «красную» ситуацию, из которой нужно выйти как можно скорее.
 - Напишите только тот код, который необходим, что бы выбраться из красной ситуации. Затем выполните рефакторинг и сделайте код красивым
 - Попытка внести слишком много изменений, пока код не компилируется, или тесты не выполняются подвергается риску усугубления проблем
 - Придерживаясь этого подхода, вы вынуждены писать небольшие тесты, подразумевающие небольшие изменения, которые помагают поддерживать работу на сложных системах