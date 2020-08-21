## backend-trainee-assignment

В релизы выложил docker-compose сборку сервиса.
Доступны следующие команды:
* `docker-compose up`
* `docker-compose stop`

При первом запуске будет произведена инициализация хранилища, поэтому в логе будет "многобуков".

### Конфиг
Если конфиг не предоставлен (файл .config), то приложение запустится с дефолтной конфигурацией, указанной ниже.
Конфиг должен находится в той же папке, что и исполняемый файл. Данные в нем указываются в формате JSON.  
Пример конфигурационного файла с настройками по умолчанию:

```
{
"port": 9000,
"ip": "",
"storage_conn_num_of_attempts": 3, // количество ретраев соединения к хранилищу, если оно недоступно
"storage_conn_interval_bw_attempts": 3, // интервал в секундах между ретраями
"log_mode": false, // логирование запросов к бд в файл log.log рядом с исполняемым файлом 
"database": {
"host": "localhost",
"port": 5432,
"user": "postgres",
"password": "123",
"name": "bta_dev"
}
}
```

### Флаги
* `-setschema`  
Этот флаг создает таблицы и связи между ними в хранилище (инициализирует его), но уничтожает существующие данные.
* `-prod`  
Этот флаг не позволяет запустить приложение без конфига.  
Также, если предоставлен этот флаг, то флаг `-setschema` будет проигнорирован.
* `-help`  
Выводит сводную информацию по имеющимся у приложения флагам.

### О реализации:
* Хранилище данных: PostgreSQL
* Паттерн: MVC
* Формат ответа:  
`{"Error":<данные>,"Result":<данные>}`  
где Error - дополнительное описание ошибки (помимо информации, получаемой из HTTP-кода ошибки).  
И Error и Result могут быть null, это означает отсутствие ошибки/результата соответственно, причем HTTP-код
ошибки может быть 200 - это значит что все в порядке, просто запрашиваемых данных нет.
* На эндпоинте /chats/get сделал так, что отсортированы не только чаты в требуемом порядке, но и сообщения в каждом из них от 
позднего к раннему. Думаю, так будет удобнее фронту.
* Лог при false (параметр "log_mode" в конфиге) пишет в stdout только ошибки от хранилища, 
при true пишет ошибки от хранилища + все запросы к нему в файл. 
Т.е. этот параметр влияет только на то, какие сообщения от хранилища будут выводиться и куда.  
Все сообщения приложения, включая ошибки, всегда идут в stdout.
* На /chats/add все дубли в "users" будут удалены молча.

### Вопросы/Предложения
1. Нужно ли на эндпоинте /chats/get подгружать все поля у вложенных сущностей? 
Например у юзера. Реализовал подгрузку только на 1 уровень вложенности.
2. На маршруте /chats/get если в чатах нет сообщений, то такие чаты будут в самом конце выборки. 
Пойдет ли такое поведение или можно придумать что-нибудь получше?
3. Не совсем хорошо знаю где какой http код применяется в случае ошибки, но оставил пока такие.
