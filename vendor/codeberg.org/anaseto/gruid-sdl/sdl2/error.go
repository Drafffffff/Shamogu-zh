package sdl

/*
#include "sdl_wrapper.h"

void GoSetError(const char *fmt) {
  SDL_SetError("%s", fmt);
}

*/
// #include "sdl_wrapper.h"
import "C"
import "errors"

var emptyCString *C.char = C.CString("")

// GetError returns the last error that occurred, or an empty string if there
// hasn't been an error message set since the last call to ClearError().
// (https://wiki.libsdl.org/SDL2/SDL_GetError)
func GetError() error {
	if err := C.SDL_GetError(); err != nil {
		gostr := C.GoString(err)
		// SDL_GetError returns "an empty string if there hasn't been an error message"
		if len(gostr) > 0 {
			return errors.New(gostr)
		}
	}
	return nil
}

// SetError set the SDL error message.
// (https://wiki.libsdl.org/SDL2/SDL_SetError)
func SetError(err error) {
	if err != nil {
		C.GoSetError(C.CString(err.Error()))
		return
	}
	C.GoSetError(emptyCString)
}

// ClearError clears any previous error message.
// (https://wiki.libsdl.org/SDL2/SDL_ClearError)
func ClearError() {
	C.SDL_ClearError()
}

// errorFromInt returns GetError() if passed negative value, otherwise it returns nil.
func errorFromInt(code int) (err error) {
	if code < 0 {
		err = GetError()
		if err == nil {
			err = errors.New("unknown error (maybe using old version of SDL2)")
		}
	}
	return
}
