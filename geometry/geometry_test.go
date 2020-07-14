package geometry

import "testing"

func TestPerimeter(t *testing.T) {
	rectangle := Rectangle{10.0, 10.0}
	got := rectangle.Perimeter()
	want := 40.0

	if got != want {
		t.Errorf("got %.2f want %.2f", got, want)
	}
}

func TestArea(t *testing.T) {
	checkArea := func(t *testing.T, shape Shape, want float64) {
		t.Helper()
		got := shape.Area()
		if got != want {
			t.Errorf("%#v got %g want %g", shape, got, want)
		}
	}

	/*t.Run("rectangles", func(t *testing.T) {
		rectangle := Rectangle{12.0, 6.0}

		checkArea(t, rectangle, 72.0)
	})

	t.Run("circles", func(t *testing.T) {
		circle := Circle{10}
		want := 314.1592653589793
		checkArea(t, circle, want)
	})

	t.Run("right triangle", func(t *testing.T) {
		triangle := RighTriangle{10.0, 5.0}
		checkArea(t, triangle, 25.)
	})*/

	areaTests := []struct {
		shape  Shape
		want   float64
		coment string
	}{
		{
			shape: Rectangle{
				Width:  12.,
				Height: 6,
			},
			want:   72.,
			coment: "rectangle",
		},
		{
			shape: RighTriangle{
				leg1: 10.,
				leg2: 5.,
			},
			want:   25.,
			coment: "right triangle",
		},
		{
			shape: Circle{
				radius: 10,
			},
			want:   314.1592653589793,
			coment: "circle",
		},
	}

	for _, tt := range areaTests {
		//checkArea(t, tt.shape, tt.want)
		t.Run(tt.coment, func(t *testing.T) {
			checkArea(t, tt.shape, tt.want)
		})
	}
}
