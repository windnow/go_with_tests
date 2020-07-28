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
/mnt/c/Users/ermek/develop/go_with_tests/build_app/internal/gameserver/server_test.go:166:11: undefined: FileSystemPlayerStore
```

## Напишем минимальное количество кода для теста необходимое для запуска провального теста и проверки его вывода

Давайте определим `FileSystemPlayerStore` в новом файле

```go
type FileSystemPlayerStore struct{}
```

## Попытайтесь снова

```
/mnt/c/Users/ermek/develop/go_with_tests/build_app/internal/gameserver/server_test.go:165:37: too many values in infsstore.FileSystemPlayerStore literal
/mnt/c/Users/ermek/develop/go_with_tests/build_app/internal/gameserver/server_test.go:167:15: store.GetLeague undefined (type infsstore.FileSystemPlayerStore has no field or method GetLeague
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