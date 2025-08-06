package main

import (
	"fmt"
)

func main() {
	// var name string = "Suwapich"
	var age int = 20

	email := "arsa_s4@silpakorn.edu"
	gpa := 3.0

	firstname, lastname := "Suwapich", "Arsa"

	fmt.Printf("Name: %s %s\nAge: %d\nEmail: %s\nGPA: %.2f\n", firstname, lastname, age, email, gpa)
}