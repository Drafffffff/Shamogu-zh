package sdl

/*
#include "sdl_wrapper.h"

*/
import "C"
import "unsafe"

// Surface contains a collection of pixels used in software blitting.
// (https://wiki.libsdl.org/SDL2/SDL_Surface)
type Surface struct {
	flags    uint32         // (internal use)
	Format   *PixelFormat   // the format of the pixels stored in the surface (read-only) (https://wiki.libsdl.org/SDL2/SDL_PixelFormat)
	W        int32          // the width in pixels (read-only)
	H        int32          // the height in pixels (read-only)
	Pitch    int32          // the length of a row of pixels in bytes (read-only)
	pixels   unsafe.Pointer // the pointer to the actual pixel data; use Pixels() for access
	UserData unsafe.Pointer // an arbitrary pointer you can set
	locked   int32          // used for surfaces that require locking (internal use)
	lockData unsafe.Pointer // used for surfaces that require locking (internal use)
	ClipRect Rect           // a Rect structure used to clip blits to the surface which can be set by SetClipRect() (read-only)
	_        unsafe.Pointer // map; info for fast blit mapping to other surfaces (internal use)
	RefCount int32          // reference count that can be incremented by the application
}

func (surface *Surface) cptr() *C.SDL_Surface {
	return (*C.SDL_Surface)(unsafe.Pointer(surface))
}

// Free frees the RGB surface.
// (https://wiki.libsdl.org/SDL2/SDL_FreeSurface)
func (surface *Surface) Free() {
	C.SDL_FreeSurface(surface.cptr())
}

// rwops provides an abstract interface to stream I/O. Applications can generally ignore the specifics of this structure's internals and treat them as opaque pointers. The details are important to lower-level code that might need to implement one of these, however.
// (https://wiki.libsdl.org/SDL2/SDL_RWops)
type rwops C.SDL_RWops

func (rwops *rwops) cptr() *C.SDL_RWops {
	return (*C.SDL_RWops)(rwops)
}

// LoadBMPRW loads a BMP image from a memory data slice.
func BMPFromMem(mem []byte) (*Surface, error) {
	if mem == nil {
		mem = []byte{}
	}
	rwops := (*rwops)(unsafe.Pointer(C.SDL_RWFromMem(unsafe.Pointer(unsafe.SliceData(mem)), C.int(len(mem)))))
	if rwops == nil {
		return nil, GetError()
	}
	surface := (*Surface)(unsafe.Pointer(C.SDL_LoadBMP_RW(rwops.cptr(), C.int(Btoi(true)))))
	if surface == nil {
		return nil, GetError()
	}
	return surface, nil
}
