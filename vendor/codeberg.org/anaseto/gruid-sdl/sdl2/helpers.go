package sdl

// Btoi returns 0 or 1 according to the value of b.
func Btoi(b bool) int {
	// Pattern optimized by Go. See Go issue 6011.
	var i int
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}
