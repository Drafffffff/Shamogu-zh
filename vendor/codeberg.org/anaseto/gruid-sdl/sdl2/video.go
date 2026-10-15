package sdl

/*
#include "sdl_wrapper.h"
#if !(SDL_VERSION_ATLEAST(2,0,5))

#if defined(WARN_OUTDATED)
#pragma message("SDL_SetWindowResizable is not supported before SDL 2.0.5")
#endif

static void SDL_SetWindowResizable(SDL_Window *window, SDL_bool resizable)
{
}

#endif
*/
import "C"
import "unsafe"

// An enumeration of window states.
// (https://wiki.libsdl.org/SDL2/SDL_WindowFlags)
const (
	WINDOW_FULLSCREEN = C.SDL_WINDOW_FULLSCREEN // fullscreen window
	WINDOW_SHOWN      = C.SDL_WINDOW_SHOWN      // window is visible
)

// An enumeration of window events.
// (https://wiki.libsdl.org/SDL2/SDL_WindowEventID)
const (
	WINDOWEVENT_NONE    = C.SDL_WINDOWEVENT_NONE    // (never used)
	WINDOWEVENT_EXPOSED = C.SDL_WINDOWEVENT_EXPOSED // window has been exposed and should be redrawn
	WINDOWEVENT_RESIZED = C.SDL_WINDOWEVENT_RESIZED // window has been resized
)

// Window position flags.
// (https://wiki.libsdl.org/SDL2/SDL_CreateWindow)
const (
	WINDOWPOS_UNDEFINED_MASK = C.SDL_WINDOWPOS_UNDEFINED_MASK // used to indicate that you don't care what the window position is
	WINDOWPOS_UNDEFINED      = C.SDL_WINDOWPOS_UNDEFINED      // used to indicate that you don't care what the window position is
	WINDOWPOS_CENTERED_MASK  = C.SDL_WINDOWPOS_CENTERED_MASK  // used to indicate that the window position should be centered
	WINDOWPOS_CENTERED       = C.SDL_WINDOWPOS_CENTERED       // used to indicate that the window position should be centered
)

// Window is a type used to identify a window.
type Window C.SDL_Window

func (window *Window) cptr() *C.SDL_Window {
	return (*C.SDL_Window)(unsafe.Pointer(window))
}

// CreateWindow creates a window with the specified position, dimensions, and flags.
// (https://wiki.libsdl.org/SDL2/SDL_CreateWindow)
func CreateWindow(title string, x, y, w, h int32, flags uint32) (*Window, error) {
	var _window = C.SDL_CreateWindow(C.CString(title), C.int(x), C.int(y), C.int(w), C.int(h), C.Uint32(flags))
	if _window == nil {
		return nil, GetError()
	}
	return (*Window)(unsafe.Pointer(_window)), nil
}

// Destroy destroys the window.
// (https://wiki.libsdl.org/SDL2/SDL_DestroyWindow)
func (window *Window) Destroy() error {
	lastErr := GetError()
	ClearError()
	C.SDL_DestroyWindow(window.cptr())
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}

// SetIcon sets the icon for the window.
// (https://wiki.libsdl.org/SDL2/SDL_SetWindowIcon)
func (window *Window) SetIcon(icon *Surface) {
	C.SDL_SetWindowIcon(window.cptr(), icon.cptr())
}

// SetResizable sets the user-resizable state of the window.
// (https://wiki.libsdl.org/SDL2/SDL_SetWindowResizable)
func (window *Window) SetResizable(resizable bool) {
	C.SDL_SetWindowResizable(window.cptr(), C.SDL_bool(Btoi(resizable)))
}

// SetSize sets the size of the window's client area.
// (https://wiki.libsdl.org/SDL2/SDL_SetWindowSize)
func (window *Window) SetSize(w, h int32) {
	C.SDL_SetWindowSize(window.cptr(), C.int(w), C.int(h))
}

// GetSize returns the size of the window's client area.
// (https://wiki.libsdl.org/SDL2/SDL_GetWindowSize)
func (window *Window) GetSize() (w, h int32) {
	var _w, _h C.int
	C.SDL_GetWindowSize(window.cptr(), &_w, &_h)
	return int32(_w), int32(_h)
}

// SetFullscreen sets the window's fullscreen state.
// (https://wiki.libsdl.org/SDL2/SDL_SetWindowFullscreen)
func (window *Window) SetFullscreen(flags uint32) error {
	return errorFromInt(int(
		C.SDL_SetWindowFullscreen(window.cptr(), C.Uint32(flags))))
}

// SetTitle sets the title of the window.
// (https://wiki.libsdl.org/SDL2/SDL_SetWindowTitle)
func (window *Window) SetTitle(title string) {
	C.SDL_SetWindowTitle(window.cptr(), C.CString(title))
}
