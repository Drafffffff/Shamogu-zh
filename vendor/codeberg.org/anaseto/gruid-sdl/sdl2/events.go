package sdl

/*
#include "sdl_wrapper.h"
#include "events.h"

// NOTE: To prevent build from failing when using older SDL2, we create a
// structure definiton that directly maps to the SDL2 struct definition if
// using the latest SDL2. Otherwise, we copy the latest definition and paste
// it here.
#if SDL_VERSION_ATLEAST(2,0,22)
typedef SDL_MouseButtonEvent MouseButtonEvent;
typedef SDL_MouseWheelEvent MouseWheelEvent;
#else
typedef struct MouseButtonEvent
{
    Uint32 type;        // ::SDL_MOUSEBUTTONDOWN or ::SDL_MOUSEBUTTONUP
    Uint32 timestamp;
    Uint32 windowID;    // The window with mouse focus, if any
    Uint32 which;       // The mouse instance id, or SDL_TOUCH_MOUSEID
    Uint8 button;       // The mouse button index
    Uint8 state;        // ::SDL_PRESSED or ::SDL_RELEASED
    Uint8 clicks;       // 1 for single-click, 2 for double-click, etc.
    Uint8 padding1;
    Sint32 x;           // X coordinate, relative to window
    Sint32 y;           // Y coordinate, relative to window
} MouseButtonEvent;

typedef struct MouseWheelEvent
{
    Uint32 type;        // ::SDL_MOUSEWHEEL
    Uint32 timestamp;   // In milliseconds, populated using SDL_GetTicks()
    Uint32 windowID;    // The window with mouse focus, if any
    Uint32 which;       // The mouse instance id, or SDL_TOUCH_MOUSEID
    Sint32 x;           // The amount scrolled horizontally, positive to the right and negative to the left
    Sint32 y;           // The amount scrolled vertically, positive away from the user and negative toward the user
    Uint32 direction;   // Set to one of the SDL_MOUSEWHEEL_* defines. When FLIPPED the values in X and Y will be opposite. Multiply by -1 to change them back
    float preciseX;     // The amount scrolled horizontally, positive to the right and negative to the left, with float precision (added in 2.0.18)
    float preciseY;     // The amount scrolled vertically, positive away from the user and negative toward the user, with float precision (added in 2.0.18)
} MouseWheelEvent;

#endif
*/
import "C"
import "unsafe"

// Enumeration of the types of events that can be delivered.
// (https://wiki.libsdl.org/SDL2/SDL_EventType)
const (
	FIRSTEVENT = C.SDL_FIRSTEVENT // do not remove (unused)

	// Application events
	QUIT = C.SDL_QUIT // user-requested quit

	// Window events
	WINDOWEVENT = C.SDL_WINDOWEVENT // window state change

	// Keyboard events
	KEYDOWN   = C.SDL_KEYDOWN   // key pressed
	KEYUP     = C.SDL_KEYUP     // key released
	TEXTINPUT = C.SDL_TEXTINPUT // keyboard text input

	// Mouse events
	MOUSEMOTION     = C.SDL_MOUSEMOTION     // mouse moved
	MOUSEBUTTONDOWN = C.SDL_MOUSEBUTTONDOWN // mouse button pressed
	MOUSEBUTTONUP   = C.SDL_MOUSEBUTTONUP   // mouse button released
	MOUSEWHEEL      = C.SDL_MOUSEWHEEL      // mouse wheel motion
)

// Toggles for different event state functions.
const (
	QUERY   = C.SDL_QUERY
	IGNORE  = C.SDL_IGNORE
	DISABLE = C.SDL_DISABLE
	ENABLE  = C.SDL_ENABLE
)

// Event is a union of all event structures used in SDL.
// (https://wiki.libsdl.org/SDL2/SDL_Event)
type Event interface {
	GetType() uint32      // GetType returns the event type
	GetTimestamp() uint32 // GetTimestamp returns the timestamp of the event
}

// CEvent is a union of all event structures used in SDL.
// (https://wiki.libsdl.org/SDL2/SDL_Event)
type CEvent struct {
	Type uint32
	_    [52]byte // padding
}

// WindowEvent contains window state change event data.
// (https://wiki.libsdl.org/SDL2/SDL_WindowEvent)
type WindowEvent struct {
	Type      uint32 // WINDOWEVENT
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the associated window
	Event     uint8  // (https://wiki.libsdl.org/SDL2/SDL_WindowEventID)
	_         uint8  // padding
	_         uint8  // padding
	_         uint8  // padding
	Data1     int32  // event dependent data
	Data2     int32  // event dependent data
}

// GetType returns the event type.
func (e *WindowEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *WindowEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// KeyboardEvent contains keyboard key down event information.
// (https://wiki.libsdl.org/SDL2/SDL_KeyboardEvent)
type KeyboardEvent struct {
	Type      uint32 // KEYDOWN, KEYUP
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with keyboard focus, if any
	State     uint8  // PRESSED, RELEASED
	Repeat    uint8  // non-zero if this is a key repeat
	_         uint8  // padding
	_         uint8  // padding
	Keysym    Keysym // Keysym representing the key that was pressed or released
}

// GetType returns the event type.
func (e *KeyboardEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *KeyboardEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// TextInputEvent contains keyboard text input event information.
// (https://wiki.libsdl.org/SDL2/SDL_TextInputEvent)
type TextInputEvent struct {
	Type      uint32                               // TEXTINPUT
	Timestamp uint32                               // timestamp of the event
	WindowID  uint32                               // the window with keyboard focus, if any
	Text      [C.SDL_TEXTINPUTEVENT_TEXT_SIZE]byte // the null-terminated input text in UTF-8 encoding
}

// GetType returns the event type.
func (e *TextInputEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *TextInputEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetText returns the text as string
func (e *TextInputEvent) GetText() string {
	length := func(buf []byte) int {
		for i := range buf {
			if buf[i] == 0 {
				return i
			}
		}

		return 0
	}(e.Text[:])

	text := e.Text[:length]
	return string(text)
}

// MouseMotionEvent contains mouse motion event information.
// (https://wiki.libsdl.org/SDL2/SDL_MouseMotionEvent)
type MouseMotionEvent struct {
	Type      uint32 // MOUSEMOTION
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with mouse focus, if any
	Which     uint32 // the mouse instance id, or TOUCH_MOUSEID
	State     uint32 // BUTTON_LEFT, BUTTON_MIDDLE, BUTTON_RIGHT, BUTTON_X1, BUTTON_X2
	X         int32  // X coordinate, relative to window
	Y         int32  // Y coordinate, relative to window
	XRel      int32  // relative motion in the X direction
	YRel      int32  // relative motion in the Y direction
}

// GetType returns the event type.
func (e *MouseMotionEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseMotionEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// MouseButtonEvent contains mouse button event information.
// (https://wiki.libsdl.org/SDL2/SDL_MouseButtonEvent)
type MouseButtonEvent struct {
	Type      uint32 // MOUSEBUTTONDOWN, MOUSEBUTTONUP
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with mouse focus, if any
	Which     uint32 // the mouse instance id, or TOUCH_MOUSEID
	Button    uint8  // BUTTON_LEFT, BUTTON_MIDDLE, BUTTON_RIGHT, BUTTON_X1, BUTTON_X2
	State     uint8  // PRESSED, RELEASED
	Clicks    uint8  // 1 for single-click, 2 for double-click, etc. (>= SDL 2.0.2)
	_         uint8  // padding
	X         int32  // X coordinate, relative to window
	Y         int32  // Y coordinate, relative to window
}

// GetType returns the event type.
func (e *MouseButtonEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseButtonEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// MouseWheelEvent contains mouse wheel event information.
// (https://wiki.libsdl.org/SDL2/SDL_MouseWheelEvent)
type MouseWheelEvent struct {
	Type      uint32  // MOUSEWHEEL
	Timestamp uint32  // timestamp of the event
	WindowID  uint32  // the window with mouse focus, if any
	Which     uint32  // the mouse instance id, or TOUCH_MOUSEID
	X         int32   // the amount scrolled horizontally, positive to the right and negative to the left
	Y         int32   // the amount scrolled vertically, positive away from the user and negative toward the user
	Direction uint32  // MOUSEWHEEL_NORMAL, MOUSEWHEEL_FLIPPED (>= SDL 2.0.4)
	PreciseX  float32 // The amount scrolled horizontally, positive to the right and negative to the left, with float precision (added in 2.0.18)
	PreciseY  float32 // The amount scrolled vertically, positive away from the user and negative toward the user, with float precision (added in 2.0.18)
}

// GetType returns the event type.
func (e *MouseWheelEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseWheelEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// QuitEvent contains the "quit requested" event.
// (https://wiki.libsdl.org/SDL2/SDL_QuitEvent)
type QuitEvent struct {
	Type      uint32 // QUIT
	Timestamp uint32 // timestamp of the event
}

// GetType returns the event type.
func (e *QuitEvent) GetType() uint32 {
	return e.Type
}

// GetTimestamp returns the timestamp of the event.
func (e *QuitEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// PollEvent polls for currently pending events.
// (https://wiki.libsdl.org/SDL2/SDL_PollEvent)
func PollEvent() Event {
	ret := C.PollEvent()
	if ret == 0 {
		return nil
	}
	return goEvent((*CEvent)(unsafe.Pointer(&C.event)))
}

func goEvent(cevent *CEvent) Event {
	switch cevent.Type {
	case WINDOWEVENT:
		return (*WindowEvent)(unsafe.Pointer(cevent))
	case KEYDOWN, KEYUP:
		return (*KeyboardEvent)(unsafe.Pointer(cevent))
	case TEXTINPUT:
		return (*TextInputEvent)(unsafe.Pointer(cevent))
	case MOUSEMOTION:
		return (*MouseMotionEvent)(unsafe.Pointer(cevent))
	case MOUSEBUTTONDOWN, MOUSEBUTTONUP:
		return (*MouseButtonEvent)(unsafe.Pointer(cevent))
	case MOUSEWHEEL:
		return (*MouseWheelEvent)(unsafe.Pointer(cevent))
	case QUIT:
		return (*QuitEvent)(unsafe.Pointer(cevent))
	default:
		return nil
	}
}
