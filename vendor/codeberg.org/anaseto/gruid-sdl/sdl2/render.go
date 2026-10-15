package sdl

/*
#include "sdl_wrapper.h"

// WORKAROUND: This prevents audio from seemingly going corrupt when drawing outside the screen bounding box?
// It does that by allocating SDL_Rect in the C context instead of Go context.
static inline int RenderCopy(SDL_Renderer *renderer, SDL_Texture *texture, SDL_Rect *src, int dst_x, int dst_y, int dst_w, int dst_h)
{
	SDL_Rect dst = {dst_x, dst_y, dst_w, dst_h};
	return SDL_RenderCopy(renderer, texture, src, &dst);
}
*/
import "C"
import (
	"unsafe"
)

// An enumeration of flags used when creating a rendering context.
// (https://wiki.libsdl.org/SDL2/SDL_RendererFlags)
const (
	RENDERER_ACCELERATED   = C.SDL_RENDERER_ACCELERATED   // the renderer uses hardware acceleration
	RENDERER_TARGETTEXTURE = C.SDL_RENDERER_TARGETTEXTURE // the renderer supports rendering to texture
)

// An enumeration of texture access patterns..
// (https://wiki.libsdl.org/SDL2/SDL_TextureAccess)
const (
	TEXTUREACCESS_TARGET = C.SDL_TEXTUREACCESS_TARGET // can be used as a render target
)

// CreateRenderer returns a new 2D rendering context for a window.
// (https://wiki.libsdl.org/SDL2/SDL_CreateRenderer)
func CreateRenderer(window *Window, index int, flags uint32) (*Renderer, error) {
	renderer := C.SDL_CreateRenderer(window.cptr(), C.int(index), C.Uint32(flags))
	if renderer == nil {
		return nil, GetError()
	}
	return (*Renderer)(unsafe.Pointer(renderer)), nil
}

// BlendMode is an enumeration of blend modes used in Render.Copy() and drawing operations.
// (https://wiki.libsdl.org/SDL2/SDL_BlendMode)
type BlendMode uint32

const (
	BLENDMODE_NONE = C.SDL_BLENDMODE_NONE // no blending
	// BLENDMODE_BLEND = C.SDL_BLENDMODE_BLEND // alpha blending
)

func (bm BlendMode) c() C.SDL_BlendMode {
	return C.SDL_BlendMode(C.Uint32(bm))
}

// CreateTexture returns a new texture for a rendering context.
// (https://wiki.libsdl.org/SDL2/SDL_CreateTexture)
func (renderer *Renderer) CreateTexture(format uint32, access int, w, h int32) (*Texture, error) {
	texture := C.SDL_CreateTexture(
		renderer.cptr(),
		C.Uint32(format),
		C.int(access),
		C.int(w),
		C.int(h))
	if texture == nil {
		return nil, GetError()
	}
	err := errorFromInt(int(C.SDL_SetTextureBlendMode(texture, BlendMode(BLENDMODE_NONE).c())))
	return (*Texture)(unsafe.Pointer(texture)), err
}

// CreateTextureFromSurface returns a new texture from an existing surface.
// (https://wiki.libsdl.org/SDL2/SDL_CreateTextureFromSurface)
func (renderer *Renderer) CreateTextureFromSurface(surface *Surface) (*Texture, error) {
	texture := C.SDL_CreateTextureFromSurface(renderer.cptr(), surface.cptr())
	if texture == nil {
		return nil, GetError()
	}
	err := errorFromInt(int(C.SDL_SetTextureBlendMode(texture, BlendMode(BLENDMODE_NONE).c())))
	return (*Texture)(unsafe.Pointer(texture)), err
}

// SetRenderTarget sets a texture as the current rendering target.
// (https://wiki.libsdl.org/SDL2/SDL_SetRenderTarget)
func (renderer *Renderer) SetRenderTarget(texture *Texture) error {
	return errorFromInt(int(
		C.SDL_SetRenderTarget(renderer.cptr(), texture.cptr())))

}

// SetScale sets the drawing scale for rendering on the current target.
// (https://wiki.libsdl.org/SDL2/SDL_RenderSetScale)
func (renderer *Renderer) SetScale(scaleX, scaleY float32) error {
	return errorFromInt(int(
		C.SDL_RenderSetScale(
			renderer.cptr(),
			C.float(scaleX),
			C.float(scaleY))))
}

// Clear clears the current rendering target with the drawing color.
// (https://wiki.libsdl.org/SDL2/SDL_RenderClear)
func (renderer *Renderer) Clear() error {
	return errorFromInt(int(
		C.SDL_RenderClear(renderer.cptr())))
}

// Copy copies a portion of the texture to the current rendering target.
// (https://wiki.libsdl.org/SDL2/SDL_RenderCopy)
func (renderer *Renderer) Copy(texture *Texture, src, dst *Rect) error {
	if dst == nil {
		return errorFromInt(int(
			C.SDL_RenderCopy(
				renderer.cptr(),
				texture.cptr(),
				src.cptr(),
				dst.cptr())))
	}
	return errorFromInt(int(
		C.RenderCopy(
			renderer.cptr(),
			texture.cptr(),
			src.cptr(),
			C.int(dst.X), C.int(dst.Y), C.int(dst.W), C.int(dst.H))))
}

// Present updates the screen with any rendering performed since the previous call.
// (https://wiki.libsdl.org/SDL2/SDL_RenderPresent)
func (renderer *Renderer) Present() {
	C.SDL_RenderPresent(renderer.cptr())
}

// Destroy destroys the specified texture.
// (https://wiki.libsdl.org/SDL2/SDL_DestroyTexture)
func (texture *Texture) Destroy() error {
	lastErr := GetError()
	ClearError()
	C.SDL_DestroyTexture(texture.cptr())
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}

// Destroy destroys the rendering context for a window and free associated textures.
// (https://wiki.libsdl.org/SDL2/SDL_DestroyRenderer)
func (renderer *Renderer) Destroy() error {
	lastErr := GetError()
	ClearError()
	C.SDL_DestroyRenderer(renderer.cptr())
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}
