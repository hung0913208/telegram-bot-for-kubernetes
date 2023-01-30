package test

import (
    "testing"
    "fmt"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/search"
)

func TestSearchSimple(t *testing.T) {
    results, _ := search.Search(nil, "sen cute ba chang")
    for _, record := range results {
        fmt.Println(record.Title)
        fmt.Println(record.Description)
    }
}


