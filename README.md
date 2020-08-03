## backend-trainee-assignment

Деплой:
1. Скомпилируйте приложение с помощью инструмента `go`.
2. Установите PostgreSQL Server и создайте базу данных с именем "bta_dev"
или с вашим кастомным именем, но нужно чтоб оно было таким же, как и в конфиге, про который будет рассказано далее.
3. Составьте конфиг приложения и положите его рядом с исполняемым файлом. Этот шаг можно пропустить, тогда приложение запустится
с дефолтной конфигурацией.
4. Один раз запустите приложение с флагом `-setschema`, это создаст таблицы, сущности и связи между ними в базе данных (но уничтожит существующие!).
5. Приложение готово к работе. Также у приложения есть еще один флаг: `-prod`. С ним приложение пишет лог запросов к бд в stdout и еще этот флаг 
не позволяет запустить приложение без конфига. Также если предоставлен этот флаг, то флаг `-setschema` будет проигнорирован.  
Еще можно запустить приложение с флагом `-help` который выведет сводную информацию по имеющимся у приложения флагам.

### Конфиг
Если конфиг не предоставлен (файл .config), то приложение запустится с дефолтными настройками, указанными ниже.
Конфиг должен находится в той же папке, что и исполняемый файл. Данные в нем указываются в формате JSON.  
Пример конфигурационного файла с настройками по умолчанию:

```
{
"port": 9000,
"database": {
"host": "localhost",
"port": 5432,
"user": "postgres",
"password": "123",
"name": "bta_dev"
}
}
```

## О реализации:
* Хранилище данных: PostgreSQL + gorm
* Паттерн: MVC
* Формат ответа:  
`{"Error":<данные>,"Result":<данные>}`  
где Error - дополнительное описание ошибки (помимо информации, получаемой из HTTP-кода ошибки).  
И Error и Result могут быть null, это означает отсутствие ошибки/результата соответственно, причем HTTP-код
ошибки может быть 200 - это значит что все в порядке, просто запрашиваемых данных нет.
* Формат запроса: id сущностей в полях запроса пишутся без "", т.е. например:
`{"chat": <CHAT_ID>}`, а не `{"chat": "<CHAT_ID>"}`.  
Решил так сделать потому, что все первичные ключи у меня числовые + это освобождает от необходимости парсинга и валидации строки.
* На эндпоинте /chats/get сделал так, что отсортированы не только чаты в требуемом порядке, но и сообщения в каждом из них от 
позднего к раннему. Думаю, так будет удобнее фронту.
* Также по ходу реализации у меня возникло множество вопросов/предложений/замечаний, некоторые из которых будут 
добавлены сюда позже. Вообще этот API можно дорабатывать и дорабатывать, но возможно Вам уже понравится такой вариант.

### Вопросы/Предложения/Обусловленности/Заметки/TODO
1. Нужно ли на эндпоинте /chats/get подгружать все поля у вложенных сущностей? 
Например у юзера. Реализовал подгрузку только на 1 уровень вложенности.
2. Нет комментариев пакетов и сущностей в них.
3. На маршруте /chats/get если в чатах нет сообщений, то такие чаты будут в самом конце выборки. 
Пойдет ли такое поведение или можно придумать что-нибудь получше?
4. Не совсем хорошо знаю где какой http код применяется в случае ошибки, но оставил пока такие.
