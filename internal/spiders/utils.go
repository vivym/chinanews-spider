package spiders

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/vivym/chinanews-spider/internal/model"
)

func parseDate(text string) (time.Time, error) {
	re := regexp.MustCompile(`(\d{4})[年.-](\d{1,2})[月.-](\d{1,2})`)
	m := re.FindStringSubmatch(text)
	if len(m) < 4 {
		return time.Now(), errors.New("invalid str: " + text)
	}

	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

func mergeTags(a []model.Tag, b []model.Tag) []model.Tag {
	var c []model.Tag

	merge := func(tags []model.Tag) {
		for _, tag := range tags {
			idx := -1
			for i, tc := range c {
				if tag.Name == tc.Name {
					idx = i
				}
			}
			if idx == -1 {
				c = append(c, tag)
			} else {
				c[idx].Count += tag.Count
			}
		}
	}

	merge(a)
	merge(b)

	return c
}
