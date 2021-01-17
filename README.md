# Тестовое задание

Связь: telegram @Aleksey_Sir

Запуск: ```docker-compose up```

Файлы xlsx таблиц находятся в ```./build/nginx/static```

Для асинхронной схемы работы необходимо выполнить Lua скрип.
[(см. Подготовка Tarantool)](#Подготовка-Tarantool)

### Результаты
  * Язык Go
  * Реализованы основные методы и методы для асинхронной работы 
  * Хранение товаров в Postgres
  * Хранение статуса задания в noSQL БД
  * Есть оценка производительности (длительность запросов, RPS)


### Добавление товаров

#### GET /products 
* seller_id, xlsx_uri передаются как GET параметры
  

* ```http://127.0.0.1:8080/products?seller_id={Id продавца}&xlsx_uri=http://nginx/{файл таблицы}```


* Пример:
 ```http://127.0.0.1:8080/products?seller_id=4&xlsx_uri=http://nginx/first_table.xlsx```
  
    Ответ: ```{"add_count":4,"update_count":0,"delete_count":0}```


### Получение товаров

#### GET /getProducts
* seller_id, offer_id, name передаются как GET параметры


* ```http://127.0.0.1:8080/getProducts?name={подстрока}&seller_id={id продавца}&offer_id={Id товара}```


* Пример:
  ```http://127.0.0.1:8080/getProducts?name=a&seller_id=1033&offer_id=15151```
  
  Ответ: ```[{"seller_id":1033,"offer_id":15151,"name":"Name_15150","price":16149,"quantity":15150,"available":true}]```


### Обновление товаров
#### Запрос на добавление и обновление одинаковый.

* Пример:
  ```http://127.0.0.1:8080/products?seller_id=4&xlsx_uri=http://nginx/first_table.xlsx```

  Ответ: ```{"add_count":0,"update_count":4,"delete_count":0}```

### Удаление товаров
#### Запрос на добавление и удаление одинаковый. За удаление товара отвечает булевое значение в таблице.

* Пример (сначала добавим, потом удалим):
  ```http://127.0.0.1:8080/products?seller_id=13&xlsx_uri=http://nginx/first_table.xlsx```
  Затем:
  ```http://127.0.0.1:8080/products?seller_id=13&xlsx_uri=http://nginx/deleteAllNotRasb.xlsx```

  Ответ: ```{"add_count":0,"update_count":1,"delete_count":3}```


### Тестирование производительности

Перед нагрузочным тестированием необходимо заполнить базу командой: ```./build/postgres/insert_update_3000000_lines.sh```
После заполнения в базе будет храниться 3 000 000 продуктов от 100 магазинов(seller_id=1000++). Если запустить
скрипт второй раз, то значения будут обновляться.


##### Таблица - 3 000 записей.

* Запрос: ```http://127.0.0.1:8080/products?seller_id=13&xlsx_uri=http://nginx/big_table.xlsx```


* Добавление: 0.39 c.
  ```{"level":"info","ts":1610811676.4648435,"msg":"/products","RequestId:":"55104dc766","Method:":"GET","RemoteAddr:":"172.18.0.1:46658","StartTime:":1610811676.0754976,"DurationTime:":0.389343419}```


* Обновление:  0.35 c.
  ```{"level":"info","ts":1610811707.584155,"msg":"/products","RequestId:":"57e9d1860d","Method:":"GET","RemoteAddr:":"172.18.0.1:46658","StartTime:":1610811707.2313178,"DurationTime:":0.35283523}```

##### Таблица - 30 000 записей.

* Запрос: ```http://127.0.0.1:8080/products?seller_id=5000&xlsx_uri=http://nginx/big_table30k.xlsx```


* Добавление: 4.79 c.
  ```{"level":"info","ts":1610811001.2551734,"msg":"/products","RequestId:":"0c97d70a13","Method:":"GET","RemoteAddr:":"172.18.0.1:41202","StartTime:":1610810996.4669857,"DurationTime:":4.788185863}```


* Обновление:  6.87 c.
  ```{"level":"info","ts":1610811042.6696548,"msg":"/products","RequestId:":"136f9ac707","Method:":"GET","RemoteAddr:":"172.18.0.1:41202","StartTime:":1610811035.8013558,"DurationTime:":6.868296862}```


### ab тесты

##### Получение записей:

* Команда: ```ab -c 150 -n 600 http://127.0.0.1:8080/getProducts?seller_id=1033```


* Результат: RPS: 55.77 (54.65 c добавлением offer_id=14000; 55.35 c еще добавлением name=a)



### Асинхронная схема работы	 
Для реализации такой схемы нам нужна была структура map[id_задания]=статус_задания.
Изначально планировал сделать мапку с мьютексами в Go, но потом решить использовать key-value БД.
Использование отдельной БД позволит нам не контролировать, чтобы запрос на получение статуса приходил бы на тот
же бэкенд, на котором пользователь получил id задания. (если мы захотим масштабировать бэк)
Еще мы не будем терять время на блокировках мьютекса.
В качестве noSQL БД я выбрал [Tarantool](https://www.tarantool.io/ru/).

#### Подготовка Tarantool

  * После запуска контейнеров выполнить ```docker exec -t -i testtask_tarantool_1 console```
  * Скопировать и выполнить скрипт из файла ```./build/tarantool/startScript.lua```
  * Если надо удалить таблицу: ```box.space.statuses:drop()```
  * Если надо посмотреть кол-во записей: ```box.space.statuses:count()```
  * Если надо посмотреть записи: ```box.space.statuses:select()```

#### Отправка задания и получение task_id

##### GET /productsAsync - отправка задания

##### GET /getStatus - просмотр статуса

### Пример асинхронной работы

##### (!!!) Для демонстрации работы я добавил sleep-ы, чтобы можно было успеть посмотреть каждый статус.
Sleep-ы добавлены на уровне логики в метод AddProductsAsync. Файл: ```./internal/pkg/product/usecase/product_usecase.go```

  * Запрос: ```http://127.0.0.1:8080/productsAsync?seller_id=13&xlsx_uri=http://nginx/first_table.xlsx```

  * Ответ: ```{"task_id":853344465}```

  * Затем запрос с полученным id (обновлять страницу/curl, смотреть смену статуса): ```http://127.0.0.1:8080/getStatus?task_id=853344465```

  * Ответ: ```{"status":"Task end: { AddCount:0  DeleteCount:0  UpdateCount:4","task_id":853344465}```


#### Что можно добавить в схему работы
При загрузке маленьких таблиц, нет необходимости в статусах (слишком частое обновление статуса) и в асинхронной работе.
При получении ссылки на файл таблицы можно выполнить HEAD запрос, чтобы узнать его размер и принять решение о запуске
задания асинхронно.

### Общие команды

Запуск: ```docker-compose up```

Наполнение базы: ```./build/postgres/insert_update_3000000_lines.sh```

Сборка после изменений: ```go fmt ./... && docker-compose build  && docker-compose up```

Подключение к PostgreSQL: ```psql -h 127.0.0.1 -p 5432 -d myService -U docker```
Пароль: docker

Подключение к Tarantool: ```docker exec -t -i testtask_tarantool_1 consol```
