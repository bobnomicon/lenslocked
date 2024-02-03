package main

import (
	"html/template"
	"os"
)

type User struct {
	Name string
	Age int
	Bio string
}

func main() {
	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}

	user := User{
		Name: "John Smith",
		Age: 123,
		Bio: "<script>alert(\"Haha, you have been h4x0r3d!\");</script>",
	}

	err = t.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
}