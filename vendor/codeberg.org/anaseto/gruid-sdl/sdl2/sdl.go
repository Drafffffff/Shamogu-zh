// Package sdl provides SDL2 bindings for gruid. It enables interoperability
// between Go and the SDL2 graphics library which is written in C. That means
// the original SDL2 installation is required for this to work.
package sdl

/*
#include "sdl_wrapper.h"

*/
import "C"

import (
	"runtime"
)

// These are the flags which may be passed to SDL_Init().
// (https://wiki.libsdl.org/SDL2/SDL_Init)
const (
	INIT_VIDEO  = C.SDL_INIT_VIDEO  // video subsystem; automatically initializes the events subsystem
	INIT_EVENTS = C.SDL_INIT_EVENTS // events subsystem
)

const (
	RELEASED = 0
	PRESSED  = 1
)

// Calls a function in the main thread. It is only properly initialized inside
// sdl.Main(..). As a default, it panics. It is used by sdl.Do(..) below.
var callInMain = func(f func()) {
	panic("sdl.Main(main func()) must be called before sdl.Do(f func())")
}

func init() {
	// Make sure the main goroutine is bound to the main thread.
	runtime.LockOSThread()
}

// Main entry point. Run this function at the beginning of main(), and pass your
// own main body to it as a function. E.g.:
//
//	func main() {
//		sdl.Main(func() {
//			// Your code here....
//			// [....]
//
//			// Calls to SDL can be made by any goroutine, but always guarded by sdl.Do()
//			sdl.Do(func() {
//				sdl.Init(0)
//			})
//		})
//	}
//
// Avoid calling functions like os.Exit(..) within your passed-in function since
// they don't respect deferred calls. Instead, do this:
//
//	func main() {
//		var exitcode int
//		sdl.Main(func() {
//			exitcode = run()) // assuming run has signature func() int
//		})
//		os.Exit(exitcode)
//	}
func Main(main func()) {
	// Queue of functions that are thread-sensitive
	callQueue := make(chan func())

	// Properly initialize callInMain for use by sdl.Do(..)
	callInMain = func(f func()) {
		done := make(chan bool, 1)
		callQueue <- func() {
			f()
			done <- true
		}
		<-done
	}

	go func() {
		main()
		// fmt.Println("END") // to check if os.Exit(..) is called by main() above
		close(callQueue)
	}()

	for f := range callQueue {
		f()
	}
}

// Do the specified function in the main thread.
// For this function to work, you must have correctly used sdl.Main(..) in your
// main() function. Calling this function before/without sdl.Main(..) will cause
// a panic.
func Do(f func()) {
	callInMain(f)
}

// Init initialize the SDL library. This must be called before using most other SDL functions.
// (https://wiki.libsdl.org/SDL2/SDL_Init)
func Init(flags uint32) error {
	if C.SDL_Init(C.Uint32(flags)) != 0 {
		return GetError()
	}
	return nil
}

// Quit cleans up all initialized subsystems. You should call it upon all exit conditions.
// (https://wiki.libsdl.org/SDL2/SDL_Quit)
func Quit() {
	C.SDL_Quit()
}
