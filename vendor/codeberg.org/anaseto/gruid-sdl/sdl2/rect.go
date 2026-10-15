package sdl

// #include "sdl_wrapper.h"
import "C"
import "unsafe"

// Point defines a two dimensional point.
// (https://wiki.libsdl.org/SDL2/SDL_Point)
type Point struct {
	X int32 // the x coordinate of the point
	Y int32 // the y coordinate of the point
}

// Rect contains the definition of a rectangle, with the origin at the upper left.
// (https://wiki.libsdl.org/SDL2/SDL_Rect)
type Rect struct {
	X int32 // the x location of the rectangle's upper left corner
	Y int32 // the y location of the rectangle's upper left corner
	W int32 // the width of the rectangle
	H int32 // the height of the rectangle
}

func (a *Rect) cptr() *C.SDL_Rect {
	return (*C.SDL_Rect)(unsafe.Pointer(a))
}
