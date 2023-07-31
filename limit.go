package where

import (
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// parseOffsetLimit 解析 offset 和 limit
func parseOffsetLimit(l *string, opt *options.FindOptions) {
	re := regexp.MustCompile(`(?i)limit[0-9,\s]+`)
	found := strings.TrimSpace(re.FindString(*l))
	if idx := strings.Index(found, " "); idx != -1 {
		*l = strings.Replace(*l, found, "", -1)
		found = strings.TrimSpace(found[idx:])
	}
	split := strings.Split(found, ",")
	if len(split) == 0 {
		return
	}

	var (
		offset int64
		limit  int64
	)
	slen := len(split)
	if slen == 1 {
		limit, _ = strconv.ParseInt(split[0], 0, 64)
	} else if slen == 2 {
		offset, _ = strconv.ParseInt(split[0], 0, 64)
		limit, _ = strconv.ParseInt(split[1], 0, 64)
	}

	if opt == nil {
		opt = options.Find()
	}
	opt = opt.SetSkip(offset)
	opt = opt.SetLimit(limit)
}
