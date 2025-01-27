package main

func main() {
	server := InitWebServer()

	err := server.Run(":8081")
	if err != nil {
		panic("start server failed")
	}
}
