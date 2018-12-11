package debug

import (
	_ "github.com/mkevac/debugcharts"
	"log"
	"net/http"
)

func StartDebug()  {
	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Println(err)
	}
}