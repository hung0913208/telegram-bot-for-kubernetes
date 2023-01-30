package test

import (
    "testing"
    "fmt"

	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/toolbox"
)

func TestParser(t *testing.T) {
    output := ""
    iface := toolbox.NewToolbox("", &output)
    iface.Execute([]string{
        "vercel",
        "env",
        "VERCEL_URL",
    })
    fmt.Println(output)
}
