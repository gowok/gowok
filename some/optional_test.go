package some

import (
	"errors"
	"testing"

	"github.com/gowok/should"
)

func TestEmpty(t *testing.T) {
	car := Empty[string]()

	should.NotNil(t, car)
	should.Nil(t, car.value)
	should.False(t, car.isPresent)
}

type Test struct {
	Description string
	Test        func(tt Test, t *testing.T)
}

func TestOf(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		input := "limo"
		car := Of(&input)

		should.NotNil(t, car)
		should.NotNil(t, car.value)
		should.True(t, car.isPresent)
		should.Equal(t, *car.value, input)
	})

	t.Run("input nil", func(t *testing.T) {
		car := Of[string](nil)

		should.NotNil(t, car)
		should.Nil(t, car.value)
		should.False(t, car.isPresent)
	})
}

func TestGet(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		input := "limo"
		car := Of(&input)

		output, ok := car.Get()

		should.True(t, ok)
		should.Equal(t, output, input)
	})

	t.Run("input nil", func(t *testing.T) {
		car := Of[string](nil)
		output, ok := car.Get()

		should.False(t, ok)
		should.Equal(t, output, "")
	})

	t.Run("empty", func(t *testing.T) {
		car := Empty[string]()
		output, ok := car.Get()

		should.False(t, ok)
		should.Equal(t, output, "")
	})
}

func TestIsPresent(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		input := "limo"
		car := Of(&input)

		isPresent := car.IsPresent()
		should.True(t, isPresent)
	})

	t.Run("input nil", func(t *testing.T) {
		car := Of[string](nil)
		isPresent := car.IsPresent()
		should.False(t, isPresent)
	})

	t.Run("empty", func(t *testing.T) {
		car := Empty[string]()
		isPresent := car.IsPresent()
		should.False(t, isPresent)
	})
}

func TestOrElse(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		input := "limo"
		car := Of(&input).OrElse("")
		should.Equal(t, car, input)
	})

	t.Run("input nil", func(t *testing.T) {
		input := "limo"
		car := Of[string](nil).OrElse(input)
		should.Equal(t, car, input)
	})

	t.Run("empty", func(t *testing.T) {
		input := "limo"
		car := Empty[string]().OrElse(input)
		should.Equal(t, car, input)
	})
}

func TestOrElseFunc(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		input := "limo"
		car := Of(&input).OrElseFunc(func() string {
			return ""
		})
		should.Equal(t, car, input)
	})

	t.Run("input nil", func(t *testing.T) {
		input := "limo"
		car := Empty[string]().OrElseFunc(func() string {
			return input
		})
		should.Equal(t, car, input)
	})
}

func TestOrPanic(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		defer func() {
			err := recover()
			should.Nil(t, err)
		}()
		input := "limo"
		Of(&input).OrPanic(errors.New("not found"))
	})

	t.Run("input nil", func(t *testing.T) {
		defer func() {
			err := recover()
			should.NotNil(t, err)
		}()
		Of[string](nil).OrPanic(errors.New("not found"))
	})

	t.Run("empty", func(t *testing.T) {
		defer func() {
			err := recover()
			should.NotNil(t, err)
		}()
		Empty[string]().OrPanic(errors.New("not found"))
	})
}
