package sdl

// #include "sdl_wrapper.h"
//
// #if !(SDL_VERSION_ATLEAST(2,0,5))
//
// enum
// {
// #if SDL_BYTEORDER == SDL_BIG_ENDIAN
//     SDL_PIXELFORMAT_RGBA32 = SDL_PIXELFORMAT_RGBA8888,
//     SDL_PIXELFORMAT_ARGB32 = SDL_PIXELFORMAT_ARGB8888,
//     SDL_PIXELFORMAT_BGRA32 = SDL_PIXELFORMAT_BGRA8888,
//     SDL_PIXELFORMAT_ABGR32 = SDL_PIXELFORMAT_ABGR8888
// #else
//     SDL_PIXELFORMAT_RGBA32 = SDL_PIXELFORMAT_ABGR8888,
//     SDL_PIXELFORMAT_ARGB32 = SDL_PIXELFORMAT_BGRA8888,
//     SDL_PIXELFORMAT_BGRA32 = SDL_PIXELFORMAT_ARGB8888,
//     SDL_PIXELFORMAT_ABGR32 = SDL_PIXELFORMAT_RGBA8888
// #endif
// };
//
// #endif
import "C"
import "image/color"

// PixelFormat contains pixel format information.
// (https://wiki.libsdl.org/SDL2/SDL_PixelFormat)
type PixelFormat struct {
	Format        uint32       // one of the PIXELFORMAT values (https://wiki.libsdl.org/SDL2/SDL_PixelFormatEnum)
	Palette       *Palette     // palette structure associated with this pixel format, or nil if the format doesn't have a palette (https://wiki.libsdl.org/SDL2/SDL_Palette)
	BitsPerPixel  uint8        // the number of significant bits in a pixel value, eg: 8, 15, 16, 24, 32
	BytesPerPixel uint8        // the number of bytes required to hold a pixel value, eg: 1, 2, 3, 4
	_             [2]uint8     // padding
	Rmask         uint32       // a mask representing the location of the red component of the pixel
	Gmask         uint32       // a mask representing the location of the green component of the pixel
	Bmask         uint32       // a mask representing the location of the blue component of the pixel
	Amask         uint32       // a mask representing the location of the alpha component of the pixel or 0 if the pixel format doesn't have any alpha information
	rLoss         uint8        // (internal use)
	gLoss         uint8        // (internal use)
	bLoss         uint8        // (internal use)
	aLoss         uint8        // (internal use)
	rShift        uint8        // (internal use)
	gShift        uint8        // (internal use)
	bShift        uint8        // (internal use)
	aShift        uint8        // (internal use)
	refCount      int32        // (internal use)
	next          *PixelFormat // (internal use)
}

// Palette contains palette information. NOTE: We don't use it in gruid.
// (https://wiki.libsdl.org/SDL2/SDL_Palette)
type Palette struct {
	Ncolors  int32  // the number of colors in the palette
	Colors   *Color // an array of Color structures representing the palette (https://wiki.libsdl.org/SDL2/SDL_Color)
	version  uint32 // incrementally tracks changes to the palette (internal use)
	refCount int32  // reference count (internal use)
}

// Color represents a color. This implements image/color.Color interface.
// (https://wiki.libsdl.org/SDL2/SDL_Color)
type Color color.NRGBA

// Pixel format values.
const (
	PIXELFORMAT_RGBA8888 = C.SDL_PIXELFORMAT_RGBA8888
	PIXELFORMAT_ABGR8888 = C.SDL_PIXELFORMAT_ABGR8888
)

// Pixel format variables.
var (
	PIXELFORMAT_RGBA32 = C.SDL_PIXELFORMAT_RGBA32
	PIXELFORMAT_ARGB32 = C.SDL_PIXELFORMAT_ARGB32
	PIXELFORMAT_BGRA32 = C.SDL_PIXELFORMAT_BGRA32
	PIXELFORMAT_ABGR32 = C.SDL_PIXELFORMAT_ABGR32
)
