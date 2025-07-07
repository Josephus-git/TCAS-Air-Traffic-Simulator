package main

import (
	"fmt"
	"os"
)

// commandExit prints a farewell message and terminates the application.
func commandExit() {
	fmt.Println("Closing TCAS-simulator... Goodbye!")
	os.Exit(0)
}
