# gobson-where

这是一个将SQL的where条件语句转换成BSON格式的Golang代码包，在使用mongoDB的时候，我们通常使用BSON格式进行条件过滤，有时候我们要写大量的代码，通过这个包你可以轻松进行条件查询。
This is a Golang code package that converts SQL WHERE condition statements to BSON format. When using MongoDB, we typically use BSON format for filtering conditions. Sometimes, we need to write a large amount of code for this purpose. With this package, you can easily perform conditional queries.


# 如何使用？

```golang
go get github.com/pywee/gobson-where
```

golang 内引入
```golang
import (
    where github.com/pywee/gobson-where
)

func main() {
   opt := where.ParseWhere(`sku!=123 AND (name=456 OR id=789) AND id!=1 LIMIT 0,10`) 
   fmt.Println(opt.Filter)
   fmt.Println(opt.Options)
}

```

The above code will eventually be transformed into the following structure:
以上的写法最终将转换为如下结构:

```
_ = bson.D{
    bson.E{Key: "sku", Value: bson.M{"$ne": "123"}},
    bson.E{Key: "$or", Value: bson.A{
        bson.D{bson.E{Key: "name", Value: 456}},
        bson.D{bson.E{Key: "id", Value: 789}},
    }},
    bson.E{Key: "id", Value: bson.M{"$ne": 1}},
}
```

针对以上的 SQL 语句中的 limit 关键词，内部同时会设定 options.FindOptions 对象
```
limit 0,10 = options.Find().SetSkip(0).SetLimit(10)
```