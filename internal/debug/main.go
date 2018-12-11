package debug

import (
	"log"
	"net/http"

	_ "github.com/mkevac/debugcharts"
)

func StartDebug() {
	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Println(err)
	}
}
