package main

// #cgo LDFLAGS: -lSDL2 -lSDL2_ttf
// #cgo pkg-config: sdl2
// #include <SDL.h>
// #include <SDL_ttf.h>
//
// SDL_EventType sdl_event_type(SDL_Event* evt) {
//     return evt->type;
// }
import "C"

import (
	"fmt"
)

func sdl_panic(failure_point string, sdl_err (*C.char)) {
	panic(failure_point + " failed: " + C.GoString(sdl_err))
}

func main() {
	if C.SDL_Init(C.SDL_INIT_EVERYTHING) != 0 {
		sdl_panic("SDL_Init", C.SDL_GetError())
	}
	defer C.SDL_Quit()
	
	if C.TTF_Init() != 0 {
		sdl_panic("TTF_Init", C.TTF_GetError())
	}
	defer C.TTF_Quit()

	win_ptr := C.SDL_CreateWindow(C.CString("This is my window"),
		C.SDL_WINDOWPOS_UNDEFINED, C.SDL_WINDOWPOS_UNDEFINED,
		800, 600,
		C.SDL_WINDOW_SHOWN)
	if win_ptr == nil {
		sdl_panic("SDL_CreateWindow", C.SDL_GetError())
	}
	defer C.SDL_DestroyWindow(win_ptr)

	// renderer_ptr := C.SDL_CreateRenderer(win_ptr, -1,
		// C.SDL_RENDERER_ACCELERATED | C.SDL_RENDERER_PRESENTVSYNC)
	// if renderer_ptr == nil {
		// sdl_panic("SDL_CreateRenderer", C.SDL_GetError())
	// }
	// defer C.SDL_DestroyRenderer(renderer_ptr)

	font_ptr := C.TTF_OpenFont(C.CString("/usr/share/fonts/truetype/noto/NotoColorEmoji.ttf"), 72)
	//font_ptr := C.TTF_OpenFont(C.CString("/usr/share/fonts/opentype/firacode/FiraCode-Regular.otf"), 24)
	if font_ptr == nil {
		sdl_panic("TTF_OpenFont", C.TTF_GetError())
	}
	defer C.TTF_CloseFont(font_ptr)
	fmt.Println("Here are", C.TTF_FontFaces(font_ptr), "faces")
	
	color := C.SDL_Color {200, 200, 200, 255}
	
	text_surf_ptr := C.TTF_RenderUTF8_Blended(font_ptr, C.CString("🙃 this => text 🙃"), color)
	if text_surf_ptr == nil {
		sdl_panic("TTF_RenderText", C.TTF_GetError())
	}
	defer C.SDL_FreeSurface(text_surf_ptr)

	// text_texture_ptr := C.SDL_CreateTextureFromSurface(renderer_ptr, text_surf_ptr)
	// if text_texture_ptr == nil {
		// sdl_panic("SDL_CreateTextureFromSurface", C.SDL_GetError())
	// }
	// defer C.SDL_DestroyTexture(text_texture_ptr)
// 
	// var width C.int
	// var height C.int
	// C.SDL_QueryTexture(text_texture_ptr, nil, nil, &width, &height)
	// //text_rect := C.SDL_Rect {0, 0, width, height}

	// C.SDL_RenderClear(renderer_ptr)
	// //C.SDL_RenderCopy(renderer_ptr, text_texture_ptr, nil, &text_rect)
	// C.SDL_RenderPresent(renderer_ptr)
	C.SDL_BlitSurface(text_surf_ptr, nil, C.SDL_GetWindowSurface(win_ptr), nil)
	// C.SDL_UpdateWindowSurface(win_ptr)
	// C.SDL_Delay(1000)

	
	running := true
	for running == true {
		var event C.SDL_Event
		for C.SDL_PollEvent(&event) != 0 {
			switch C.sdl_event_type(&event) {
				case C.SDL_QUIT:
					running = false
			}
		}

		C.SDL_UpdateWindowSurface(win_ptr)
		C.SDL_Delay(16)
	}
}
