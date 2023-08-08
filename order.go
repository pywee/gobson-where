package where

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 解析 Order 语句，当前只支持一个字段的解析
// 支持 order by createAt desc
// 不支持 order by createAt desc, id desc
func parseOrder(l *string, opt *options.FindOptions) {
	re1 := regexp.MustCompile(`\s+`)
	re2 := regexp.MustCompile(`(?i)order by [0-9a-zA-Z_]+\s+(desc|asc)`)
	found := strings.TrimSpace(re2.FindString(*l))
	result := re1.ReplaceAllString(found, " ")
	*l = strings.TrimSpace(strings.Replace(*l, found, "", -1))

	// fmt.Println(result)
	// fmt.Println(found, "xxxxxxxx")
	// fmt.Println(*l, "...")

	arr := strings.Split(result, " ")
	if len(arr) != 4 {
		return
	}
	if opt == nil {
		opt = options.Find()
	}

	var sort int8 = -1
	if strings.ToLower(arr[3]) == "asc" {
		sort = 1
	}
	opt.SetSort(bson.D{{arr[2], sort}})
}
