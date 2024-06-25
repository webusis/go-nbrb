# go-nbrb
db_connect

        User:   "username"
        Passwd: "password"
        
main

        ListenAndServe ":8888"

Возвращаемые результаты в json из консоли или http сервера

go run main.go date 431 2024-03-05 2024-06-10 ----http://localhost:8888/date/431/2024-02-05/2024-03-06---- Выборка в интервале с указанием валюты

go run main.go today 431 ----http://localhost:8888/today/431---- Выборка за сегодня с указанием валюты 

go run main.go tsync ----http://localhost:8888/tsync---- Выборка за сегодня доступных валют 

go run main.go currencies ----http://localhost:8888/currencies---- Выборка всех валют 
