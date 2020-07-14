package geometry

import "math"

// MyInt just for check
type MyInt int

// Shape Donald Duck
type Shape interface {
	Area() float64
}

// Rectangle struct ..
type Rectangle struct {
	Width  float64
	Height float64
}

// Area return area
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter return perimeter
func (r Rectangle) Perimeter() float64 {
	var two MyInt = 2
	return float64(two) * (r.Width + r.Height)
}

// RighTriangle righ triangle
type RighTriangle struct {
	leg1 float64
	leg2 float64
}

// Area return area of right triangle
func (r RighTriangle) Area() float64 {
	return r.leg1 * r.leg2 / 2
}

// Circle struct ...
type Circle struct {
	radius float64
}

// Area Return area of cirle
func (c Circle) Area() float64 {
	return math.Pi * c.radius * c.radius
}
