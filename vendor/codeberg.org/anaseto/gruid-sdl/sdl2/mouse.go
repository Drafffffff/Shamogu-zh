package sdl

/*
#include "sdl_wrapper.h"
*/
import "C"

// Used as a mask when testing buttons in buttonstate.
const (
	BUTTON_LEFT   = C.SDL_BUTTON_LEFT   // left mouse button
	BUTTON_MIDDLE = C.SDL_BUTTON_MIDDLE // middle mouse button
	BUTTON_RIGHT  = C.SDL_BUTTON_RIGHT  // right mouse button
)
