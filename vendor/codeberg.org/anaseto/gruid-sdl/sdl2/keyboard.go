package sdl

/*
#include "sdl_wrapper.h"
static void disableAutomaticScreenKeyboard(void) {
 SDL_SetHintWithPriority("SDL_ENABLE_SCREEN_KEYBOARD", "0", SDL_HINT_OVERRIDE);
 SDL_SetHintWithPriority("SDL_ENABLE_STEAM_SCREEN_KEYBOARD", "0", SDL_HINT_OVERRIDE);
}

*/
import "C"

func DisableAutomaticScreenKeyboard() { C.disableAutomaticScreenKeyboard() }

// Scancode is an SDL keyboard scancode representation.
// NOTE: unused by gruid, but it's used for Keysym struct definition.
// (https://wiki.libsdl.org/SDL2/SDL_Scancode)
type Scancode uint32

// Keysym contains key information used in key events.
// (https://wiki.libsdl.org/SDL2/SDL_Keysym)
type Keysym struct {
	Scancode Scancode // SDL physical key code
	Sym      Keycode  // SDL virtual key code
	Mod      uint16   // current key modifiers
	unused   uint32   // unused
}

// GetModState returns the current key modifier state for the keyboard.
// (https://wiki.libsdl.org/SDL2/SDL_GetModState)
func GetModState() Keymod {
	return (Keymod)(C.SDL_GetModState())
}

// StartTextInput starts accepting Unicode text input events.
// (https://wiki.libsdl.org/SDL2/SDL_StartTextInput)
func StartTextInput() {
	C.SDL_StartTextInput()
}

// StopTextInput stops receiving any text input events.
// (https://wiki.libsdl.org/SDL2/SDL_StopTextInput)
func StopTextInput() {
	C.SDL_StopTextInput()
}

// SetTextInputRect sets the rectangle used to type Unicode text inputs.
// (https://wiki.libsdl.org/SDL2/SDL_SetTextInputRect)
func SetTextInputRect(rect *Rect) {
	C.SDL_SetTextInputRect(rect.cptr())
}
