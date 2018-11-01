package xtesting_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tjamet/xgo/xtesting"
)

func TestNoEnv(t *testing.T) {
	assert.NoError(t, os.Setenv("SomeVariable", "hello-world"))
	assert.NoError(t, os.Setenv("SomeOtherVariable", "hello-world"))
	assert.NoError(t, os.Setenv("Variable", "hello=world"))
	t.Run("NoEnv removes the variable from environment", func(t *testing.T) {
		defer xtesting.NoEnv("Some*")()
		assert.Equal(t, "", os.Getenv("SomeVariable"))
		assert.Equal(t, "", os.Getenv("SomeOtherVariable"))
	})
	t.Run("Environment variable is reset at the end of the test", func(t *testing.T) {
		assert.Equal(t, "hello-world", os.Getenv("SomeVariable"))
		assert.Equal(t, "hello-world", os.Getenv("SomeOtherVariable"))
	})
}

func TestInEnv(t *testing.T) {
	t.Run("InEnv adds environment variable", func(t *testing.T) {
		defer xtesting.InEnv("AVariable", "hello-world")()
		assert.Equal(t, "hello-world", os.Getenv("AVariable"))
	})
	t.Run("After the test, the environment variable is removed", func(t *testing.T) {
		assert.Equal(t, "", os.Getenv("AVariable"))
	})
}

func ExampleNoEnv() {
	os.Setenv("SomeVariable", "value1")
	os.Setenv("SomeOtherVariable", "value2")

	func() {
		defer xtesting.NoEnv("Some*")()
		fmt.Printf("in function:%s\n", os.Getenv("SomeVariable"))
	}()
	fmt.Println("out of function:", os.Getenv("SomeVariable"))
	// Output:
	// in function:
	// out of function: value1
}

func ExampleInEnv() {
	func() {
		defer xtesting.InEnv("VariableName", "value1")()
		fmt.Printf("in function: %s\n", os.Getenv("VariableName"))
	}()
	fmt.Println("out of function:", os.Getenv("VariableName"))
	// Output:
	// in function: value1
	// out of function:
}
