package where

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
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

func Parse(conditions string) bson.D {
	if len(strings.TrimSpace(conditions)) == 0 {
		return bson.D{{}}
	}

	var (
		k     int8
		where = make(map[string]*bson.D, 1)
	)

	conditions = strings.Replace(conditions, " and ", " AND ", -1)
	conditions = strings.Replace(conditions, " or ", " OR ", -1)
	str := []rune(conditions)
	sql := make([]rune, 0, len(str))

	if !strings.Contains(conditions, "deleted") {
		conditions += ` AND deleted=0 `
	}

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
					tmp := Parse(string(kstr[1 : len(kstr)-1]))
					key := fmt.Sprintf(`$%d`, k)
					where[key] = &tmp
					sql = append(sql, []rune(key)...)
					break
				}
			}
			continue
		}
		sql = append(sql, str[i])
	}

	// fmt.Println(string(sql))
	// d = append(d, parseAndOr(string(sql))...)

	return parseAndOr(string(sql), where)
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

	filter = bson.E{
		Key:   strings.TrimSpace(cds[:idx]),
		Value: bson.M{syn: strings.TrimSpace(strings.Replace(cds[idx+step:], `"`, "", 2))},
	}
	return filter
}

// parseAndOr 递归解析AND/OR
func parseAndOr(conditions string, where map[string]*bson.D) bson.D {
	var cs bson.D
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
