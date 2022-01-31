package main

import (
	"fmt"
	"net/http"
)

func webpageHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("webpageHandler called\n")
	http.ServeFile(w, r, "static/webpage-admin/index.html")
}

func main() {

	fmt.Printf("Jabuti X firmware mainpoint")
	fmt.Printf("teste")

	InitRpio()

	//MoveForward()
	//MoveLeft()
	//MoveRight()

	//return

	var defaultPort string = "3000"

	var port string = defaultPort

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/admin", webpageHandler)

	http.HandleFunc("/api", ApiHandler)

	fmt.Printf("Jabuti X is listening to port 3000\n")
	http.ListenAndServe(":"+port, nil)
}
