package where

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Parse 递归解析括号
// 括号为最高优先级
// a=1 AND ((c!=2) and d=9) OR a=3

// 用法:
/*
	filter := where.Parse(`sku!=123 AND (name=3 or sku=5)`)
	以上写法最终得到的BSON结构是:
	bson.D{
		bson.E{Key: "sku", Value: bson.M{"$ne": "123"}},
		bson.E{Key: "$or", Value: bson.A{
			bson.D{bson.E{Key: "name", Value: 4}},
			bson.D{bson.E{Key: "name", Value: 5}},
		}},
	}
*/

type opts struct {
	Filter  bson.D
	Options *options.FindOptions
}

// Parse 解析SQL句子
// 包含 Where、Limit、Offset 关键字
func Parse(conditions string, params ...interface{}) *opts {
	var realParams = make([]interface{}, 0, len(params))
	for _, v := range params {
		if !strings.Contains(conditions, "?") {
			break
		}
		kind := reflect.TypeOf(v).Kind().String()
		if kind == "string" {
			conditions = strings.Replace(conditions, "?", `"%s"`, 1)
			realParams = append(realParams, v)
		} else if strings.Contains(kind, "int") {
			conditions = strings.Replace(conditions, "?", `%d`, 1)
			realParams = append(realParams, v)
		} else if strings.Contains(kind, "float") {
			conditions = strings.Replace(conditions, "?", "%f", 1)
			realParams = append(realParams, v)
		}
	}

	if conditions = strings.TrimSpace(conditions); len(conditions) == 0 {
		conditions = "deleted!=1"
	}

	if len(realParams) > 0 {
		conditions = fmt.Sprintf(conditions, realParams...)
	}
	var (
		k     int8
		where = make(map[string]*bson.D, 1)
	)

	opt := options.Find()
	parseOrder(&conditions, opt)
	parseOffsetLimit(&conditions, opt)

	conditions = strings.Replace(conditions, " and ", " AND ", -1)
	conditions = strings.Replace(conditions, " or ", " OR ", -1)

	if !strings.Contains(conditions, "deleted") {
		if len(conditions) == 0 {
			conditions = `deleted!=1`
		} else {
			conditions += ` AND deleted!=1 `
		}
	}

	str := []rune(conditions)
	sql := make([]rune, 0, len(str))

	for i := 0; i < len(str); i++ {
		if str[i] == '(' {
			k++
			step := 0
			kstr := make([]rune, 0, 30)
			for j := i; j < len(str); j++ {
				kstr = append(kstr, str[j])
				if str[j] == '(' {
					step++
				} else if str[j] == ')' {
					step--
				}
				if step == 0 {
					i = j
					// d = append(d, Parse(string(kstr[1:len(kstr)-1]))...)
					// fmt.Println(string(kstr[1:len(kstr)-1]), "::")
					opt2 := Parse(string(kstr[1 : len(kstr)-1]))
					key := fmt.Sprintf(`$%d`, k)
					where[key] = &opt2.Filter
					sql = append(sql, []rune(key)...)
					break
				}
			}
			continue
		}
		sql = append(sql, str[i])
	}

	// fmt.Println(string(sql), "...")
	// d = append(d, parseAndOr(string(sql))...)
	return &opts{
		Filter:  parseAndOr(string(sql), where),
		Options: opt,
	}
}

// parseWhereSymbool 解析符号
func parseWhereSymbool(cds string, where map[string]*bson.D) bson.E {
	var filter bson.E
	syn := "="
	idx := -1
	step := 1
	if idx = strings.Index(cds, "!="); idx != -1 {
		step = 2
		syn = "$ne"
	} else if idx = strings.Index(cds, ">="); idx != -1 {
		step = 2
		syn = "$gte"
	} else if idx = strings.Index(cds, "<="); idx != -1 {
		step = 2
		syn = "$lte"
	} else if idx = strings.Index(cds, "="); idx != -1 {
		syn = "$eq"
	} else if idx = strings.Index(cds, ">"); idx != -1 {
		syn = "$gt"
	} else if idx = strings.Index(cds, "<"); idx != -1 {
		syn = "$lt"
	}

	column := strings.TrimSpace(cds[:idx])
	if column == "id" {
		column = "_id"
	}
	filter = bson.E{
		Key: strings.TrimSpace(column),
	}

	value := strings.TrimSpace(cds[idx+step:])
	if strings.Count(value, `"`) >= 2 {
		thisValue := strings.TrimSpace(strings.Replace(value, `"`, "", -1))
		if column == "_id" {
			oid, _ := primitive.ObjectIDFromHex(thisValue)
			filter.Value = bson.M{syn: oid}
		} else {
			filter.Value = bson.M{syn: thisValue}
		}
	} else if strings.Contains(value, ".") {
		valueFloat, _ := strconv.ParseFloat(value, 64)
		filter.Value = bson.M{syn: valueFloat}
	} else {
		valueInt, _ := strconv.ParseInt(value, 10, 64)
		filter.Value = bson.M{syn: valueInt}
	}

	return filter
}

// parseAndOr 递归解析AND/OR
func parseAndOr(conditions string, where map[string]*bson.D) bson.D {
	var cs = bson.D{}
	if idx := strings.Index(conditions, "OR"); idx != -1 {
		cs = append(cs, bson.E{
			Key: "$or",
			Value: bson.A{
				parseAndOr(strings.TrimSpace(conditions[:idx]), where),
				parseAndOr(strings.TrimSpace(conditions[idx+2:]), where),
			},
		})
	} else if idx := strings.Index(conditions, "AND"); idx != -1 {
		cs = append(cs, bson.E{
			Key: "$and",
			Value: bson.A{
				parseAndOr(strings.TrimSpace(conditions[:idx]), where),
				parseAndOr(strings.TrimSpace(conditions[idx+3:]), where),
			},
		})
	} else {
		if strings.Contains(conditions, "$") {
			cs = append(cs, *(where[conditions])...)
		} else {
			cs = append(cs, parseWhereSymbool(strings.TrimSpace(conditions), where))
		}
	}
	return cs
}
