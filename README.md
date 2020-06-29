# Конвертор alfa2ynab
Позволяет сконвертировать CSV-выгрузку платежных операций Альфа-банка в подходящий для импорта в [YNAB](https://www.youneedabudget.com/) фомат.

## Установка
```
go get -u github.com/scripter-v/alfa2ynab
```

## Использование
```
alfa2ynab < ~/Downloads/movementList.csv > alfa_ops.csv
```
