package main

import (
	"errors"
	"fmt"
)

type Student struct {
	ID string `json:"id"` 
	Name string `json:"name"`
	Email string `json:"email"`
	Year int `json:"year"`
	GPA float64 `json:"gpa"`
}

func (s *Student) IsHonor() bool {
	return s.GPA >= 3.50
}

func (s *Student) Validate() error {
	if s.Name == "" {
		return errors.New("Name is required")
	}
	if s.Year < 1 || s.Year > 4 {
		return errors.New("Year must be between 1-4")
	}
	if s.GPA < 0 || s.GPA > 4 {
		return errors.New("GPA must be between 0-4")
	}
	return nil
} 

func main() {
	// var st Student = Student{ID:"1", Name:"Suwapich", Email:"arsa_s4@silpakorn.edu", Year:3, GPA:3.0} 

	// st := Student{ID:"1", Name:"Suwapich", Email:"arsa_s4@silpakorn.edu", Year:3, GPA:3.0} 

	students := []Student{
		{ID:"1", Name:"A", Email:"A@silpakorn.edu", Year:1, GPA:1.0},
		{ID:"2", Name:"B", Email:"B@silpakorn.edu", Year:2, GPA:2.0},
		{ID:"3", Name:"C", Email:"C@silpakorn.edu", Year:3, GPA:3.0},
	}

	newStudent := Student{ID:"4", Name:"D", Email:"D@silpakorn.edu", Year:4, GPA:4.0}
	students = append(students, newStudent)

	for i, student := range students {
		fmt.Printf("Student %d: Honor  %v\n", i, student.IsHonor())
		fmt.Printf("Student %d: Validation  %v\n", i, student.Validate())
	}
}