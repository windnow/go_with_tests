Ввод/Вывод и сортировка
-----------------------

В предыдущей главе мы провели очередную итерацию нашего приложения добавив новый эндпойнт `/league`. В процессе мы научились работать с JSON, встраивать типы и управлять маршрутизацией

Заказчик обеспокоен тем, что продукт теряет оценки при перезапуске сервера. Это потому, что хранилище реализовано в памяти. Он так же не доволен тем, что в `/league` мы не реализовали возврат икроков, отсортированных по количеству побед

Продолжим!

## Хранение данных

Существует десятки баз данных, которые мы могли бы использовать для этого, однако мы применем очень простой подход. Мы собираемся хранить данные приложения в файле формата JSON. 

Это позволит данным быть весьма переносимыми и относительно просто в реализации.

Такая реализация не будет достаточно хорошо масштабироваться, но, учитывая что мы работаем с прототипом, это примемлемо. При изменении обстоятельств, при которых решение будет не приемлемо, будет просто заменить на что то другое, реализующее интерфейс `PlayerStore`, который мы используем.

Пока мы сохраним `InMemoryPlayerStore`, что бы интеграционные тесты продолжали проходить при разработке нового хранилища. Как только мы убедимся, что новая реализация достаточно функциональна, мы поменяем `InMemoryPlayerStore` и удалим её

## Сначал напишем тест

Сейчас вы уже должны быть знакомы с интерфейсам стандартной библиотеки для чтения (`io.Reader`) и записи (`io.Writer`) данных а так же, как, используя стандартную библиотеку для теста этих функций, не использовать реальные файлы.

Для завершения работы нам необходимо реализовать `PlayerStore`, поэтому напишем тесты для хранилища, обращающиеся к методам, которые должны быть реализованы. Давайте начнем с `GetLeague`.

```go
func TestFileSystemStore(t *testing.T) {
	t.Run("/league from a reader", func(t *testing.T) {
		database := strings.NewReader(`[
			{"Name": "Cleo", "Wins": 10},
			{"Name": "Chris", "Wins": 33}]`)

	    store := FileSystemPlayerStore{database}

    	got := store.GetLeague()

    	want := []gs.Player{
    		{"Cleo", 10},
    		{"Chris", 33},
    	}

    	assertLeague(t, got, want)
	})
}
```

Мы используем `strings.NewReader`, который вернет нам `Reader`, который будет использовать наш `FileSystemPlayerStore` для чтения данных. В `main` мы откроем файл, который так же является `Reader`'ом

## Попытайтесь запустить тест

```
github.com/windnow/edusrv/internal/gameserver_test [github.com/windnow/edusrv/internal/gameserver.test]
server_test.go:166:11: undefined: FileSystemPlayerStore
```

## Напишем минимальное количество кода для теста необходимое для запуска провального теста и проверки его вывода

Давайте определим `FileSystemPlayerStore` в новом файле

```go
type FileSystemPlayerStore struct{}
```

## Попытайтесь снова

```
server_test.go:165:37: too many values in infsstore.FileSystemPlayerStore literal
server_test.go:167:15: store.GetLeague undefined (type infsstore.FileSystemPlayerStore has no field or method GetLeague
```

Ошибк возникает, поскольку мы передаем `Reader`, но структура не ожидает его, а так же не определен метод `GetLeague`.

```go
type FileSystemPlayerStore struct {
	database io.Reader
}

// GetLeague ...
func (f *FileSystemPlayerStore) GetLeague() []gs.Player {
	return nil
}

// NewFileSystemPlayerStore моя реализация, позволяющая задавать скрытые поля, 
// недоступные извне
func NewFileSystemPlayerStore(database io.Reader) *FileSystemPlayerStore {
	return &FileSystemPlayerStore{database}
}
```

Еще одна попытка... 

```
--- FAIL: TestFileSystemStore (0.00s)
    --- FAIL: TestFileSystemStore//league_from_a_reader (0.00s)
        server_test.go:174: got [], want [{Cleo 10} {Chris 33}]
```

## Напишем минимальный код для прохождения теста

Ранее мы уже читали JSON

```go
func (f *FileSystemPlayerStore) GetLeague() []gs.Player {
	var league []gs.Player
	json.NewDecoder(f.database).Decode(&league)
	return league
}
```

Тест должен пройти.

## Рефакторинг

Это уже сделано! Код нашего теста для сервера должен декодировать JSON из запроса.

Давайте воспользуемся DRY и вынесем это в функцию

Создайте новый файл с именем `league.go` и добавьте в него следующий код:

```go
func NewLeague(rdr io.Reader) ([]gs.Player, error) {
	var league []gs.Player
	err := json.NewDecoder(rdr).Decode(&league)
	if err != nil {
		err = fmt.Errorf("problem parsing league, %v", err)
	}

	return league, err
}
```

Вызывайте ее из нашей реализации файлового хранилища, а так же из тестового хелпера `getLeagueFromResponse` в файле `server_test.go`

```go
func (f *FileSystemPlayerStore) GetLeague() []gs.Player {
	league, _ := gs.NewLeague(f.database)
	return league
}
```

Пока у нас нет стратегии работы с ошибками, но давайте продолжим

## Поиск проблем

В нашей реализации есть изъян. Во первых, давайте вспомним, как устроен `io.Reader`:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Можно представить, что наш файл читается побайтова до конча. Что же произойдет, если попытаться прочитать во второй раз?

Добавьте следующее в конец текущего теста

```go
		got = store.GetLeague()
		assertLeague(t, got, want)
```

Мы ожидаем, что тест пройдет, однако это не так.

Проблема в том, что `Reader` достиг конца файла, и ему больше нечего читать. Необходимо как то сообщить ему, что нужно вернуться в начало

[`ReaderSeeker`](https://golang.org/pkg/io/#ReadSeeker) еще один интерфейс из стандартной библиотеки, который может помочь

```go
type ReaderSeeker interface {
    Reader
    Seeker
}
```

Помните встраивание? этот интерфейс состоит из интефейсов `Reader` и [`Seeker`](https://golang.org/pkg/io/#Seeker)

```go
type Seeker interface {
    Seek(offset int64, whence int) (int64, error)
}
```

Выглядит хорошо, можем ли мы использовать этот интефейс в `FileSystemPlayerStore`?

```go
type FileSystemPlayerStore struct {
	database io.ReadSeeker
}

func (f *FileSystemPlayerStore) GetLeague() []gs.Player {
	f.database.Seek(0, 0)
	league, _ := gs.NewLeague(f.database)
	return league
}
```

Запустив тест, вы увидите, что он пройдет! К нашему счастью `strings.NewReadr`, использованный нами в тестах, так же реализует `ReadSeeker`, так что не нужно производить больше никаких изменений

Далее мы реализуем `GetPlayerScore`

## Сначала напишем тест

```go
	t.Run("get player score", func(t *testing.T) {
		database := strings.NewReader(`[
			{"Name": "Cleo", "Wins": 10},
			{"Name": "Chris", "Wins": 33}]`)

		store := fs.NewFileSystemPlayerStore(database)

		got := store.GetPlayerScore("Chris")

		want := 33

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
```

## Попробуйте запустить тест
```
server_test.go:186:15: store.GetPlayerScore undefined (type *infsstore.FileSystemPlayerStore has no field or method GetPlayerScore)
```

## Напишем минимум кода необходимого для запуска теста и проверим сообщение об ошибки теста

Для компиляции теста необходимо добавить новый метод к нашему новому типу

```go
func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {
	return 0
}
```

Теперь он скомпилируется, но тест провалится.

```
--- FAIL: TestFileSystemStore (0.00s)
    --- FAIL: TestFileSystemStore/get_player_score (0.00s)
        server_test.go:191: got 0, want 33
```

## Напишем немного кода для прохождения теста

Мы можем запустить цикл по всей лиге для поиска игрока и вернуть его счет

```go
func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {
	var wins int = 0

	for _, player := range f.GetLeague() {
		if player.Name == name {
			wins = player.Wins
			break
		}
	}
	return wins
}
```

## Рефакторинг

Вы уже виделм множество переносов проверок в хелперы, поэтому следующий приведенный код сделайте рабочим.

```go
	t.Run("/get player score", func(t *testing.T) {
		database := strings.NewReader(`[
			{"Name": "Cleo", "Wins": 10},
			{"Name": "Chris", "Wins": 33}]`)

		store := fs.NewFileSystemPlayerStore(database)

		got := store.GetPlayerScore("Chris")
		want := 33
		assertScoreEquals(t, got, want)
	})
```

В завершении нузно реализовать запись счета с помощью `RecordWin`

-----------------------------------------------------------------

## Сначала напишем тест

Наш подход достаточно непрактичен для записи. Мы не можем просто обновить одну запись JSON в файле. Нам нужно будет записывать полностью новое представление базы данных каждый раз.

Как мы запишем? Обычно мы использовали `Writer`, но у нас есть `ReadSeeker`. Теоретически мы могли бы использовать два интерфейса, однако, стандартная библиотека уже имеет для нас интефейс `ReadWriteSeeker`, позволяющий нам производить с файлом все необходимые операции.

Давайте обновим тип

```go
type FileSystemPlayerStore struct {
	database io.ReadWriteSeeker
}
```

При компиляции: 

```
/infsstore.go:36:32: cannot use database (type io.ReadSeeker) as type io.ReadWriteSeeker in field value:
	io.ReadSeeker does not implement io.ReadWriteSeeker (missing Write method)
```

Не удивительно, поскольку `strings.Reader` не реализует `ReadWriteSeeker`, но что же нам делать?

У нас есть два варианта:

 - Создать временный файл для кадого теста. `*os.File` реализует `ReadWriteSeeker`. Преимущество в том, что тест станет скорее интеграционным, поскольку мы действительно читаем и пишем из файловой системы, поэтому к тесту будет более высокий уровень доверия. Минусы в том, что мы используем модульные тесты по причине их большей скорости и простоты. Так же нам добавится работы в целях создания временного файла а так же проверки того, что он удалится после теста.

  - Мы моем использовать стороннюю библиотеку. [Matetti](https://github.com/mattetti) написал библиотеку [filebuffer](https://github.com/mattetti/filebuffer), которая реализует необходимый нам интерфейс без задействования файловой системы

я не думаю, что оба варианты приемлемы, однако, выбрав вариант со сторонней библиотекой, мне пришлось бы объяснять управление зависимостями :) поэтому будем использовать файлы.

Перед добавлением нового, нужно скомпилировать имеющиеся тесты заменив `strings.Reader` на `os.File`

Давайте создадим вспомогательную функцию которая будет создавать временный файл, содержащий некоторые данные

```go
func createTempFile(t *testing.T, initialData string) (io.ReadWriteSeeker, func()) {
	t.Helper()

	tmpfile, err := ioutil.TempFile("", "db")

	if err != nil {
		t.Fatalf("could't create tmp file %v", err)
	}

	tmpfile.Write([]byte(initialData))

	fmt.Println(fmt.Sprintf("created temp file %s", tmpfile.Name()))
	return tmpfile, func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}
}
```

[TempFile](https://golang.org/pkg/io/ioutil/#TempDir) создает временный файл, который мы можем использовать. Значение `"db"` устанавливается префиксом к имени файла, остальная часть которого генерируется случайно. Это обеспечивает отсутствие конфликта с другим файлом.

Заметьте, что возвращается не только `ReadWriteSeeker` (файл) а так же функция. Нам нужно быть уверенным, что файл после будет удален после завершения теста. Нам бы не хотелось, что бы информация о файле передавалась в сам тест, поскольку она подвержена ошибкам не не так интересна читателю. Возвращая функцию удаления файла, мы можем быть уверенны, что детали реализации останутся в хелпере, и все, что нужно сделать вызываемой функции это запустить `defer clearDatabase()`, т.е. вызвать функцию, возвращаемой хелпером.

```go
	t.Run("/league from a reader", func(t *testing.T) {
		database, clearDatabase := createTempFile(t, `[
			{"Name": "Cleo", "Wins": 10},
			{"Name": "Chris", "Wins": 33}]`)
		defer clearDatabase()

		store := fs.NewFileSystemPlayerStore(database)

		got := store.GetLeague()

		want := []gs.Player{
			{"Cleo", 10},
			{"Chris", 33},
		}

		assertLeague(t, got, want)

		got = store.GetLeague()
		assertLeague(t, got, want)
	})

	t.Run("/get player score", func(t *testing.T) {
		database, clearDatabase := createTempFile(t, `[
			{"Name": "Cleo", "Wins": 10},
			{"Name": "Chris", "Wins": 33}]`)
		defer clearDatabase()
		store := fs.NewFileSystemPlayerStore(database)

		got := store.GetPlayerScore("Chris")
		want := 33
		assertScoreEquals(t, got, want)
	})
```

Если запустить тесты, то они должны пройти. Произведено достаточное количество изменений но теперь кажется, что завершение определения интерфейса завершено, так что будет легко добавить новые тесты

Давайте полчим первую итерацию записи победы для существующего игрока

```go
	t.Run("store wins for existing players", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Cleo", "Wins" 10},
			{"Name": "Chris", "Wins" 33}
		]`)

		defer cleanDatabase()

		store := fs.NewFileSystemPlayerStore(database)

		store.RecordWin("Chris")

		got := store.GetPlayerScore("Chris")
		want := 33
		assertScoreEquals(t, got, want)
	})
```

## Попытайтесь запустить тест

```
server_test.go:204:8: store.RecordWin undefined (type *infsstore.FileSystemPlayerStore has no field or method RecordWin)
```

## Напишем минимальное количество кода для запуска теста и проверки сообщения об ошибке

Добавте новый метод

```go
func (f *FileSystemPlayerStore) RecordWin(name string){

}
```

```
--- FAIL: TestFileSystemStore (0.00s)
    --- FAIL: TestFileSystemStore/store_wins_for_existing_players (0.00s)
        server_test.go:208: got 0, want 34
```

Наша реализация пуста, и поэтому возвращено старое значение.

## Напишем немного кода для его прохождения

```go
func (f *FileSystemPlayerStore) RecordWin(name string) {
	league := f.GetLeague()

	for i, player := range league {
		if player.Name == name {
			league[i].Wins++
		}
	}
	f.database.Seek(0, 0)
	json.NewEncoder(f.database).Encode(league)
}
```

Возможно, вы задаетесь вопросом, почему вместо `player.Wins++` я использовал `league[i].Wins++`?

При проходе с помощью `range` по срезу, возвращается текущий индекс цикла (в нашем случае `i`) и *копия* элемента по этому индексу. Изменение поля `Wins` в копии не повлияет на серз `league`, по которму идет итерация. По этой причине нам нужно получить ссылку на актуальное значение с помощью `league[i]` и затем изменть его значение.

Если запустить тесты, то они должны пройти. 

## Рефакторинг

В `GetPlayerScore` и `RecordWin` мы проходим по `[]Player` для поиска игрока по имени. Нам нужно отредактировать этот общий код внутри `FileSystemStore` но мне кажется что это достаточно используемый код, который можно поднять в новый тип. Работа с "лигой" производилась с помощью `[]Player` однако, мы можем добавить новый тип под названием `League`

В `league.go` добавьте следующее

```go
type League []Player

func (l League) Find(name string) *Player {
	for i, p := range l {
		if p.Name == name {
			return &l[i]
		}
	}
	return nil
}
```

Теперь, если `League` используется где то, то там можно с легкостью производить поиск. 

Измените наш интерфейс `PlayerStore`, что бы возвращалось `League` вместо `[]Player`. При попытке перезапустить тесты, вы получите ошибки компиляции, поскольку интерфейс был изменен, однако это легко исправить; просто измените тип возвращаемого значения с `[]Player` на  `League`.

Это позволит упростить наши методы в `FileSystemStore`

```go
func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {

	player := f.GetLeague().Find(name)

	if player != nil {
		return player.Wins
	}

	return 0
}

func (f *FileSystemPlayerStore) RecordWin(name string) {
	league := f.GetLeague()
	player := league.Find(name)

	if player != nil {
		player.Wins++
	}

	f.database.Seek(0, 0)
	json.NewEncoder(f.database).Encode(league)
}
```

Так выглядит намного лучше, и теперь для `League` можно найти и другие полезные функции при реорганизации её

Теперь нужно обработать сценарий записи победы для нового игрока

## Сначала напишем тест

```go
t.Run("store wins for new players", func(t *testing.T) {
	database, cleanDatabase := createTempFile(t, `[
		{"Name": "Cleo", "Wins": 10},
		{"Name": "Chris", "Wins": 33}]`)

	defer cleanDatabase()

    store := fs.NewFileSystemPlayerStore(database)

	store.RecordWin("Pepper")
	got := store.GetPlayerScore("Pepper")
	want := 1
	assertScoreEquals(t, got, want)
})
```

## Попытайтесь запустить тест

```
--- FAIL: TestFileSystemStore/store_wins_for_new_players (0.00s)
        server_test.go:224: got 0, want 1
```

## Напишем немного кода для его прохождения

Нам нужно просто обработать сценарий, при котором `Find` возвращает `nil`, поскольку в этом случае игрок не найден

```go
func (f *FileSystemPlayerStore) RecordWin(name string) {
	league := f.GetLeague()
	player := league.Find(name)

	if player != nil {
		player.Wins++
	} else {
		league = append(league, gs.Player{Name: name, Wins: 1})
	}

	f.database.Seek(0, 0)
	json.NewEncoder(f.database).Encode(league)
}
```

Чтож, теперь мы можем попытаться использовать наше новое хранилище в интеграционном тесте. Это придаст нам уверенности, что программа работает, после чего мы можем избавиться от лишнего `InMemoryPlayerStore`

В `TestRecordingWinsAndRetrievengThem` избавтесь от старого хранилища

```go
	database, cleanDatabase := createTempFile(t, "")
	defer cleanDatabase()
	store := fs.NewFileSystemPlayerStore(database)
```

Если запустить тесты, они должны пройти и теперь мы можем удалить `InMemoryPlayerStore`. Однако в `main.go` возникнут проблемы компиляции, который мотивирует нас использовать теперь новое хранилище в "реальном" коде

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/windnow/edusrv/internal/gameserver"
	"github.com/windnow/edusrv/internal/infsstore"
)

const (
	dbFileName = "game.db.json"
)

func main() {
	db, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("problem opening %s %v", dbFileName, err)
	}

	store := infsstore.NewFileSystemPlayerStore(db)
	server := gameserver.NewServer(store)

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}

```

 - Мы создаем файл для нашей базы данных
 - Второй аргумент в `os.OpenFile` позволяет определить доступ для открытия файла. В нашем случае `O_RDWR` значит файл будет открыт для чтения и записи, а `os.O_CREATE` значит создать файл, если его нет.
 - Третий аргумент определяет права для файла. В нашем случае все пользователи могут читать и записывать файл. (см [superuser.com](https://superuser.com/questions/295591/what-is-the-meaning-of-chmod-666) для деталей)


Теперь все данные сохраняются в файле между запусками. УРА!

## Больше рефакторинга и проблемы производительности

Каждый раз при вызове `GetLeague()` или `GetPlayerScore()` мы читаем весь файл и преобразуем его в JSON. Мы не должны этого делать, поскольку `FileSystemStore` целиком ответственна за состояние лиги. Все, что нужно сделать, это прочитать файл при старте программы а так же обновить файл при изменении данных.

Мы можем создать конструктор, который может производить некоторую инициализацию, и хранить лигу как значение Для чтения в `FileSystemStore`

```go
type FileSystemPlayerStore struct {
	database io.ReadWriteSeeker
	league   gs.League
}

func NewFileSystemPlayerStore(database io.ReadWriteSeeker) *FileSystemPlayerStore {
	database.Seek(0, 0)
	league, _ := gs.NewLeague(database)
	return &FileSystemPlayerStore{
		database: database,
		league:   league,
	}
}
```

Это позволит считывать с диска лишь единожды. Теперь мы можем заменить все вызовы получения лиги с диска на `f.league`

```go
func (f *FileSystemPlayerStore) GetLeague() gs.League {
	return f.league
}

func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {

	player := f.league.Find(name)

	if player != nil {
		return player.Wins
	}

	return 0
}

// RecordWin ...
func (f *FileSystemPlayerStore) RecordWin(name string) {
	player := f.league.Find(name)

	if player != nil {
		player.Wins++
	} else {
		f.league = append(f.league, gs.Player{Name: name, Wins: 1})
	}

	f.database.Seek(0, 0)
	json.NewEncoder(f.database).Encode(f.league)
}
```

Если попробывать запустить тест, то он будет жаловаться на инициализацию `FileSystemPlayerStore`. Исправить это можно вызовом нового конструктора, вместо создания объекта на месте

    windnow:
	В моем варианте ошибки не будет, т.к. я изначально использовал конструктор

## Другие проблемы

В том, как мы работаем с файлом есть некоторая наивность, которая может создать очень неприятную проблему в будущем.

При обращении к `RecordWin` мы с помощью `Seek` идем к началу файла и затем записываем новые данные. Но что будет, если окажется, что новые данные стали меньше предыдущих?

В текущих условиях это не важно. Мы не редактируем и не удаляем счет, так что данные могут только увеличиваться. Однако было бы безответственно оставить код как есть. Не исключено, что сценарий с удалением может появится.

Как мы будем его тестировать? Все, что нам нужно сделать, это сначала отделить вид данных, которые мы пишим от собственно записи. Затем мы можем протестировать отдельно, что бы убедиться, что все работает так, как мы надеемся.

Мы создадим новый тип, инкапсулирующий функционал "перезаписываем все с начала". Я назвал его `Tape`. Создайте файл со следующим содержимым

```go
package tape

import "io"

// Tape ...
type Tape struct {
	File io.ReadWriteSeeker
}

func (t *Tape) Write(p []byte) (n int, err error) {
	t.File.Seek(0, 0)
	return t.File.Write(p)
}
```

Обратите внимание, что мы реализуем лишь `Write`, инкупсулирующая `Seek`, так что теперь `FileSystemStore` может иметь ссылку лишь на `Writer`.

```go
type FileSystemPlayerStore struct {
	database io.Writer
	league   gs.League
}
```

Обновите конструктор для использования `Type`

```go
func NewFileSystemPlayerStore(database io.ReadWriteSeeker) *FileSystemPlayerStore {
	database.Seek(0, 0)
	league, _ := gs.NewLeague(database)

	return &FileSystemPlayerStore{
		database: &tape.Tape{database},
		league:   league,
	}
}
```

В завершении мы можем получить дополнительный бонус, удалив вызов `Seek` из `RecordWin`. Да, это не очень впечетляет, но хотя бы это значит, если мы производим любую другой вид записи, мы можем надеятся что поведение нашего `Write` то, которое нам нужно. Так же это позволит нам протестировать и исправить потенциально проблемный код отдельно.

Давайте напишем тест, в котором мы обновим оригинальное содержимое чем то меньшим по объему


## Сначала напишем тест

Наш тест будет создавать файл с некоторым содержимым, пытаться записать в него с использоватьем `Tape`, и затем читать его снова для получения содержимого файла. В `tape_test.go`:

```go
package tape

import (
	"io/ioutil"
	"testing"

	. "github.com/windnow/edusrv/internal/helpers"
)

func TestTape_Write(t *testing.T) {
	file, clean := CreateTempFile(t, "12345")
	defer clean()

	tape := &Tape{file}

	tape.Write([]byte("abc"))
	file.Seek(0, 0)
	newFileContent, _ := ioutil.ReadAll(file)

	got := string(newFileContent)
	want := "abc"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

## Попытайтесь запустить тест

```
=== RUN   TestTape_Write
    TestTape_Write: tape_test.go:24: got "abc45", want "abc"
--- FAIL: TestTape_Write (0.00s)
```

Как мы и думали, необходимые данные записываются, однако остается часть старых данных


## Напишем немного кода для его прохождения

`os.File` имеет функцию усечения, позволяющая эффективно очистить файл. Все, что нужно сделать, это вызвать её.

Измените `Tape`:

```go
type Tape struct {
	File *os.File
}

func (t *Tape) Write(p []byte) (n int, err error) {
	t.File.Truncate(0)
	t.File.Seek(0, 0)
	return t.File.Write(p)
}
```

Компилятор выдаст ошибку в нескольких местах, где ожидается `io.ReadWriteSeeker`, но мы используем `*os.File`. Вы должны быть в состоянии исправить эти проблемы, однако если это вызовет проблемы, просто посмотрите код источника

После изменения `TestTape_Write` тесты должны пройти

## Еще одна небольшая корректировка

В `RecordWin` у нас есть строка `json.NewEncoder(f.database).Encode(f.league)`.

Нам не нужно каждый раз создавать новый энкодер при каждой записи. Его можно инициализировать в нашем конструкторе и использовать при записи.

Добавьте ссылку на `Encoder` в наш тип

```go
type FileSystemPlayerStore struct {
	database *json.Encoder
	league   gs.League
}
```

Инициализируйте в конструкторе

```go
func NewFileSystemPlayerStore(file *os.File) *FileSystemPlayerStore {
	file.Seek(0, 0)
	league, _ := gs.NewLeague(file)

	return &FileSystemPlayerStore{
		database: json.NewEncoder(&tape.Tape{file}),
		league:   league,
	}
}
```

И задействуйте её в `RecordWin`

## Разве мы не нарушили некоторые правила? Тестирование приватных полей? Нет интерфейсов?

### При тестировании приватных типов

Действительно, в общем следует избегать тестирования приватных полей и методов, поскольку иногда это может привести к слишком тесному связыванию реализации с тестами, что усложнит рефакторинг в будущем.

Однако, следует помнить, что тесты должны вселят *уверенность*

Мы были не уверены, что наша реализация будет работать при добавлении функций редактирования или удаления, и нам не хотелось оставлять код в таком виде. Особенно, если бы над кодом работало несколько человек, которые могли не знать о недостатках первоначального кода.

В конце концов, это лишь один тест. Если мы решим изменить работу, не составит больших проблем просто удалить его, но мы, по крайней мере, учли требования будущих сопровождающих.

### Интефейсы

Мы начали код с `io.Reader`, который был простейшим способом модульного тестирования нашего `PlayerStore`. По мере работы над кодом, мы задействовали `io.ReadWriter` и затем `io.ReadWriteSeeker`. Затем мы обнаружили, что ничего, кроме `os.File` из стандартной бибилиотеки не реализует его. Мы могли бы написать свою реализацию, однако эффективнее для тестов использовать временные файлы.

В завершении, нам понадобилось усекать, что так же реализовано `*os.File`. Можно было бы создать собственный интерфейс, отвечающий этим требованиям

```go
type ReadWriteSeekTruncate interface {
	io.ReadWriteSeeker
	Truncate(size int64) error
}
```

Однако что это нам даст? Помните, что мы не имитируем, и нереально для хранилища *файловой системы* использовать другой тип помимо `*os.File`, поэтому нам не нужен предоставляемый интерфейсами полиморфизм.

Не бойтесь эксперементировать и изменять типы подобно тому, как мы делаем это здесь. Самое замечательное в использовании статически типизированного языка - то, что компилятор придет на помощь при каждом изменении

## Обработка ошибок

Перед тем, как мы начнем работать над сортировкой, мы должны убедиться, что текущий код нас полностью устраивает, и избавится от технического долга, который у нас может быть. Это важный принцип - получать работающее ПО как можно скорее (находясь за пределами красной зоны), однако это не значит, что можно игнорировать случаи ошибок!

Вернемся к `NewFileSystemPlayerStore`, где есть `league, _ := gs.NewLeague(file)`

`NewLeague` может вернуть ошибку, если не сможет прочитать лигу из предоставленного `io.Reader`

Тогда было практично, игнорировать возможную ошибку, поскольку тогда проваливались тесты. Если бы мы попытались исправить это сразу, то мы бы пытались справится с двумя вещами одновременно. 

Давайте добавим нашему конструктору возможность возвращения ошибки

```go
func NewFileSystemPlayerStore(file *os.File) (*FileSystemPlayerStore, error) {
	file.Seek(0, 0)
	league, err := gs.NewLeague(file)

	if err != nil {
		return nil, fmt.Errorf("problem loading player store from file %s, %v", file.Name(), err)
	}

	return &FileSystemPlayerStore{
		database: json.NewEncoder(&tape.Tape{
			File: file,
		}),
		league: league,
	}, nil
}
```

Помните, что очень важно предоставлять полезные сообщения об ошибках (как ваши тесты). Люди в интернетах шутят, что большая часть кода на Go это:

```go
if err != nil {
	return err
}
```

**Это на 100% не характерно**. Добавление контекстной информации (т.е. вызвавшие ошибку действия) в сообщение об ошибке делает управление вашим приложением более легким.

При попытке компиляции вы получите некоторые ошибки

```
cmd/gamelogger/main.go:23:8: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
internal/gameserver/server_test.go:107:8: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
internal/gameserver/server_test.go:167:9: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
internal/gameserver/server_test.go:187:9: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
internal/gameserver/server_test.go:199:9: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
internal/gameserver/server_test.go:218:9: assignment mismatch: 1 variable but infsstore.NewFileSystemPlayerStore returns 2 values
```

В `main` мы выдем из программы, распечатав ошибку

```go
store, err := infsstore.NewFileSystemPlayerStore(db)
if err != nil {
	log.Fatalf("problem creating file system player store, %v", err)
}
```

В тестах мы должны проверить утверждения отсутствия ошибки. Для этого можно добавить хелпер.

```go
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("didn't expect an error but got one, %v", err)
	}
}
```

Исправьте другие ошибки компиляции с использованием этого хелпера. В конце у вас должна остаться провальный тест

```
=== RUN   TestRecordingWinsAndRetrievingThem
    TestRecordingWinsAndRetrievingThem: server_test.go:108: didn't expect an error but got one, problem loading player store from file /tmp/db511593973, problem parsing league, EOF
--- FAIL: TestRecordingWinsAndRetrievingThem (0.00s)
```

Мы не можем парсить лигу, поскольку файл пуст. Ранее мы не получали ошибку, посколку игнорировали их.

Давайте исправим наш большой интеграционный тест, добавив в него некоторый валидный JSON

```go
database, cleanDatabase := CreateTempFile(t, "[]")
// и т.д.
```

Теперь, когда все тесты пройдены, нам нужно обработать сценарий, при котором файл пуст.


## Сначала напишем тест

```go
t.Run("works with an empty file", func(t *testing.T) {
	database, clearDB := CreateTempFile(t, "")
	defer clearDB()

	_, err := fs.NewFileSystemPlayerStore(database)

	assertNoError(t, err)
})
```

## Попробуйте запустить тест

```
=== RUN   TestFileSystemStore/works_with_an_empty_file
    TestFileSystemStore/works_with_an_empty_file: server_test.go:239: didn't expect an error but got one, problem loading player store from file /tmp/db123986730, problem parsing league, EOF
```

## Напишем минимум кода для прохождения теста

Измените наш конструктор на следующее

```go
func NewFileSystemPlayerStore(file *os.File) (*FileSystemPlayerStore, error) {
	file.Seek(0, 0)
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("problem getting file info from file %v, %s", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}

	league, err := gs.NewLeague(file)

	if err != nil {
		return nil, fmt.Errorf("problem loading player store from file %s, %v", file.Name(), err)
	}

	return &FileSystemPlayerStore{
		database: json.NewEncoder(&tape.Tape{
			File: file,
		}),
		league: league,
	}, nil
}
```


`file.Stat` возвращает статистику для нашего файла, позволяющий посмотреть размер файла. Если он пуст, мы записываем в него пустой массив JSON, и возвращаемся (`Seek`) в начало, готовыми для остальной части кода.

## Рефакторинг

Конструктор сейчас стал слегка запутан, поэтому давайте извлечем инициализирующий код в отдельную функцию

```go
func initialisePlayerDBFile(file *os.File) error {
	file.Seek(0, 0)
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("problem getting file info from file %v, %s", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}
	return nil
}
```

```go
func NewFileSystemPlayerStore(file *os.File) (*FileSystemPlayerStore, error) {

	err := initialisePlayerDBFile(file)
	if err != nil {
		return nil, fmt.Errorf("problem initialising player db file, %v", err)
	}

	league, err := gs.NewLeague(file)

	if err != nil {
		return nil, fmt.Errorf("problem loading player store from file %s, %v", file.Name(), err)
	}

	return &FileSystemPlayerStore{
		database: json.NewEncoder(&tape.Tape{
			File: file,
		}),
		league: league,
	}, nil
}
```

## Сортировка

Заказчик желает, что бы `/league` возвращало игроков, отсортированных по их очкам - от больших к меньшим.

Здесь следует принять главное решение - где именно в программном коде это должно происходить. В случае, если бы мы использовали реальную БД то приминили бы `ORDER BY`, поэтому сортировка выполнилась бы очень быстро. Посему кажется, что `PlayerStore` должен отвечать за это


## Напишем сначала тест

```go
t.Run("league sorted", func(t *testing.T) {
	db, cleanDB := CreateTempFile(t, `[
		{"Name": "Cleo", "Wins": 10},
		{"Name": "Chris", "Wins": 33}]`)
	defer cleanDB()

	store, err := fs.NewFileSystemPlayerStore(db)

	assertNoError(t, err)

	got := store.GetLeague()
	want := []gs.Player{
		{"Chris", 33},
		{"Cleo", 10},
	}

	assertLeague(t, got, want)

	// read again
	got = store.GetLeague()
	assertLeague(t, got, want)
})
```

Порядок во входящем JSON файле нарушен, f наш `want` будет проверять, что возвращается верный порядок.

## Попробуйте запустить тест

```
=== RUN   TestLeague/league_sorted
    TestLeague/league_sorted: server_test.go:176: got [{Cleo 10} {Chris 33}], want [{Chris 33} {Cleo 10}]
    TestLeague/league_sorted: server_test.go:180: got [{Cleo 10} {Chris 33}], want [{Chris 33} {Cleo 10}]
```

## Напишем минимум кода для его прохождения

```go
func (f *FileSystemPlayerStore) GetLeague() gs.League {
	sort.Slice(f.league, func(i, j int) bool {
		return f.league[i].Wins > f.league[j].Wins
	})
	return f.league
}
```

[sort.Slice](https://golang.org/pkg/sort/#Slice)

    Сортирует предоставленный срез предоставленной функцией сравнения

Легко!

## Итоги

### Мы рассмотрели
 - Интерфейс `Seeker` и его отношение к `Reader` и `Writer`
 - Работу с файлами
 - Создание простого в использовании хелпера для тестов с файлами, скрывающий реализацию
 - `sort.Slice` для сортировки срезов
 - Использовани компилятора для безопасного внесения изменений в приложение.

### Нарушение правил

  - Большинство правил в разработке ПО на самом деле не правила, а просто лучшие практики, работающие в 80% случаях
  - Мы нашли ситуацию, в которой одно из наших прошлых «правил» - не тестировать внутренние функции нам не помогло. Поэтом мы его нарушили
  - При нарушении правил важно помнить о компромисе, на который вы идёте. В нашем случае, все в порядке, поскольку это всего лишь тест, однако в противном случае было бы очень сложно выполнить сценарий
  - Что бы иметь возможность нарушать правила, вы **должны сначала понять их**. Аналогия - игра на гитаре. Неважно, насколько вы креативны, вы должны понимать и практиковать основы.


## На каком этапе наше приложение?

 - У нас есть HTTP API, позволяющее создавать игроков и увеличивать их счет
 - Мы можем вернуть лигу, с очками всех в JSON виде
 - Данные сохраняются в JSON файле