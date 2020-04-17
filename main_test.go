package main

import "testing"

func TestCollisionRect(t *testing.T) {
	type Case struct {
		name                           string
		x1, y1, w1, h1, x2, y2, w2, h2 int
		expectX, expectY               int
	}

	cases := []Case{
		Case{
			"not collision",
			0, 0, 5, 5,
			10, 10, 5, 5,
			0, 0,
		},

		Case{
			"collision x",
			0, 0, 5, 5,
			4, 0, 5, 5,
			-1, -5,
		},

		Case{
			"collision x2",
			0, 0, 5, 5,
			-4, 0, 5, 5,
			1, -5,
		},

		Case{
			"collision y",
			0, 0, 5, 5,
			0, 4, 5, 5,
			-5, -1,
		},

		Case{
			"collision y2",
			0, 0, 5, 5,
			0, -4, 5, 5,
			-5, 1,
		},

		Case{
			"real case",
			15, 120, 12, 12,
			0, 112, 16, 16,
			1, 8,
		},
	}

	for _, c := range cases {
		actualX, actualY := CollideRect(c.x1, c.y1, c.w1, c.h1, c.x2, c.y2, c.w2, c.h2)
		if actualX != c.expectX || actualY != c.expectY {
			t.Errorf("%s: expect:(%d, %d) actual:(%d, %d)", c.name, c.expectX, c.expectY, actualX, actualY)
		}
	}
}
