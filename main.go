// main.go

package main

func main() {
	a := App{}
	a.Initialize(
		"test",
		"test",
		"test")

	a.Run(":8010")
}
