package sdl

// #include "sdl_wrapper.h"
import "C"

// The SDL virtual key representation.
// (https://wiki.libsdl.org/SDL2/SDL_Keycode)
// (https://wiki.libsdl.org/SDL2/SDLKeycodeLookup)
const (
	K_UNKNOWN = C.SDLK_UNKNOWN // "" (no name, empty string)

	K_RETURN    = C.SDLK_RETURN    // "Return" (the Enter key (main keyboard))
	K_ESCAPE    = C.SDLK_ESCAPE    // "Escape" (the Esc key)
	K_BACKSPACE = C.SDLK_BACKSPACE // "Backspace"
	K_TAB       = C.SDLK_TAB       // "Tab" (the Tab key)
	K_SPACE     = C.SDLK_SPACE     // "Space" (the Space Bar key(s))

	K_INSERT   = C.SDLK_INSERT   // "Insert" (insert on PC, help on some Mac keyboards (but does send code 73, not 117))
	K_HOME     = C.SDLK_HOME     // "Home"
	K_PAGEUP   = C.SDLK_PAGEUP   // "PageUp"
	K_DELETE   = C.SDLK_DELETE   // "Delete"
	K_END      = C.SDLK_END      // "End"
	K_PAGEDOWN = C.SDLK_PAGEDOWN // "PageDown"
	K_RIGHT    = C.SDLK_RIGHT    // "Right" (the Right arrow key (navigation keypad))
	K_LEFT     = C.SDLK_LEFT     // "Left" (the Left arrow key (navigation keypad))
	K_DOWN     = C.SDLK_DOWN     // "Down" (the Down arrow key (navigation keypad))
	K_UP       = C.SDLK_UP       // "Up" (the Up arrow key (navigation keypad))

	K_KP_ENTER     = C.SDLK_KP_ENTER     // "Keypad Enter" (the Enter key (numeric keypad))
	K_KP_1         = C.SDLK_KP_1         // "Keypad 1" (the 1 key (numeric keypad))
	K_KP_2         = C.SDLK_KP_2         // "Keypad 2" (the 2 key (numeric keypad))
	K_KP_3         = C.SDLK_KP_3         // "Keypad 3" (the 3 key (numeric keypad))
	K_KP_4         = C.SDLK_KP_4         // "Keypad 4" (the 4 key (numeric keypad))
	K_KP_5         = C.SDLK_KP_5         // "Keypad 5" (the 5 key (numeric keypad))
	K_KP_6         = C.SDLK_KP_6         // "Keypad 6" (the 6 key (numeric keypad))
	K_KP_7         = C.SDLK_KP_7         // "Keypad 7" (the 7 key (numeric keypad))
	K_KP_8         = C.SDLK_KP_8         // "Keypad 8" (the 8 key (numeric keypad))
	K_KP_9         = C.SDLK_KP_9         // "Keypad 9" (the 9 key (numeric keypad))
	K_KP_0         = C.SDLK_KP_0         // "Keypad 0" (the 0 key (numeric keypad))
	K_KP_PERIOD    = C.SDLK_KP_PERIOD    // "Keypad ." (the . key (numeric keypad))
	K_KP_BACKSPACE = C.SDLK_KP_BACKSPACE // "Keypad Backspace" (the Backspace key (numeric keypad))

	K_LCTRL  = C.SDLK_LCTRL  // "Left Ctrl"
	K_LSHIFT = C.SDLK_LSHIFT // "Left Shift"
	K_LALT   = C.SDLK_LALT   // "Left Alt" (alt, option)
	K_LGUI   = C.SDLK_LGUI   // "Left GUI" (windows, command (apple), meta)
	K_RCTRL  = C.SDLK_RCTRL  // "Right Ctrl"
	K_RSHIFT = C.SDLK_RSHIFT // "Right Shift"
	K_RALT   = C.SDLK_RALT   // "Right Alt" (alt, option)
	K_RGUI   = C.SDLK_RGUI   // "Right GUI" (windows, command (apple), meta)
)

// An enumeration of key modifier masks.
// (https://wiki.libsdl.org/SDL2/SDL_Keymod)
const (
	KMOD_NONE     = C.KMOD_NONE     // 0 (no modifier is applicable)
	KMOD_LSHIFT   = C.KMOD_LSHIFT   // the left Shift key is down
	KMOD_RSHIFT   = C.KMOD_RSHIFT   // the right Shift key is down
	KMOD_LCTRL    = C.KMOD_LCTRL    // the left Ctrl (Control) key is down
	KMOD_RCTRL    = C.KMOD_RCTRL    // the right Ctrl (Control) key is down
	KMOD_LALT     = C.KMOD_LALT     // the left Alt key is down
	KMOD_RALT     = C.KMOD_RALT     // the right Alt key is down
	KMOD_LGUI     = C.KMOD_LGUI     // the left GUI key (often the Windows key) is down
	KMOD_RGUI     = C.KMOD_RGUI     // the right GUI key (often the Windows key) is down
	KMOD_NUM      = C.KMOD_NUM      // the Num Lock key (may be located on an extended keypad) is down
	KMOD_CAPS     = C.KMOD_CAPS     // the Caps Lock key is down
	KMOD_MODE     = C.KMOD_MODE     // the AltGr key is down
	KMOD_CTRL     = C.KMOD_CTRL     // (KMOD_LCTRL|KMOD_RCTRL)
	KMOD_SHIFT    = C.KMOD_SHIFT    // (KMOD_LSHIFT|KMOD_RSHIFT)
	KMOD_ALT      = C.KMOD_ALT      // (KMOD_LALT|KMOD_RALT)
	KMOD_GUI      = C.KMOD_GUI      // (KMOD_LGUI|KMOD_RGUI)
	KMOD_RESERVED = C.KMOD_RESERVED // reserved for future use
)

// Keycode is the SDL virtual key representation.
// (https://wiki.libsdl.org/SDL2/SDL_Keycode)
type Keycode C.SDL_Keycode

// Keymod is a key modifier masks.
// (https://wiki.libsdl.org/SDL2/SDL_Keymod)
type Keymod C.SDL_Keymod
