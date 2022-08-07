package utils

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	//"sort"
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/models"
)

func SortSliceByRFC3339(data []models.Withdrawn) []models.Withdrawn {
	sort.Slice(data, func(i, j int) bool {
		t1 := parseStrtoRFC33339(data[i].ProcessedAt)
		t2 := parseStrtoRFC33339(data[j].ProcessedAt)
		return t1.Before(t2)
	})
	return data
}

func parseStrtoRFC33339(timeStr string) time.Time {
	//layout := "2006-01-02T15:04:05.00"
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Println(err)
	}
	return t
}

func CheckURL(conn string) bool {
	resp, err := http.Get("conn")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	log.Println(fmt.Sprint(resp.StatusCode) + resp.Status)
	return true
}
