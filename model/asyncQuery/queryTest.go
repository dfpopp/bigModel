package asyncQuery

import (
	"fmt"
	"github.com/dfpopp/bigModel"
)

func ChatTest() {
	id := "37971744365697737-8374400315131467404"
	resp, err := ChatQuery("8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe", id)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(bigModel.Json_encode(resp))
}
func VideoTest() {
	id := "37971744365697737-8374400108972951686"
	resp, err := VideoQuery("8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe", id)
	if err != nil {
		panic(err)
	}
	fmt.Println(bigModel.Json_encode(resp))
}
