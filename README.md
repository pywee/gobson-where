# gobson-where

这是一个将SQL的where条件语句转换成BSON格式的Golang代码包，在使用mongoDB的时候，我们通常使用BSON格式进行条件过滤，有时候我们要写大量的代码，通过这个包你可以轻松进行条件查询。
This is a Golang code package that converts SQL WHERE condition statements to BSON format. When using MongoDB, we typically use BSON format for filtering conditions. Sometimes, we need to write a large amount of code for this purpose. With this package, you can easily perform conditional queries.


# 如何使用？

```
go get github.com/pywee/gobson-where
```

golang 内引入
```
import (
    w github.com/pywee/gobson-where
)

func main() {
   filter := where.ParseWhere(`sku=123 AND (name=456 OR id=789)`) 
   fmt.Println(filter)
}

```