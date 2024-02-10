package main

import (
	"html/template"
	"os"
)

type User struct {
	Name string
	Age int
	Bio string
	Skills map[string]string
}

func main() {
	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}

	user := User{
		Name: "John Smith",
		Age: 123,
		Bio: "",
		Skills: map[string]string{
			"A": "JS",
			"B": "PHP",
			"C": "Python",
		},
	}

	err = t.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
}