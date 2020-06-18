package main

/*
#cgo LDFLAGS: -lSDL2 -lharfbuzz -lfreetype
#cgo pkg-config: sdl2 harfbuzz freetype2

#include <SDL.h>
#include <hb.h>
#include <hb-ft.h>

SDL_EventType sdl_event_type(SDL_Event* evt) {
	return evt->type;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func sdl_panic(failure_point string, sdl_err (*C.char)) {
	panic(failure_point + " failed: " + C.GoString(sdl_err))
}

func main() {
	if C.SDL_Init(C.SDL_INIT_EVERYTHING) != 0 {
		sdl_panic("SDL_Init", C.SDL_GetError())
	}
	defer C.SDL_Quit()

	win_ptr := C.SDL_CreateWindow(C.CString("This is my window"),
		C.SDL_WINDOWPOS_UNDEFINED, C.SDL_WINDOWPOS_UNDEFINED,
		800, 600,
		C.SDL_WINDOW_SHOWN)
	if win_ptr == nil {
		sdl_panic("SDL_CreateWindow", C.SDL_GetError())
	}
	defer C.SDL_DestroyWindow(win_ptr)
	win_surf_ptr := C.SDL_GetWindowSurface(win_ptr)

	// renderer_ptr := C.SDL_CreateRenderer(win_ptr, -1, C.SDL_RENDERER_ACCELERATED | C.SDL_RENDERER_PRESENTVSYNC)
	// if renderer_ptr == nil {
		// sdl_panic("SDL_CreateRenderer", C.SDL_GetError())
	// }
	// defer C.SDL_DestroyRenderer(renderer_ptr)


	var ft_lib C.FT_Library
	if C.FT_Init_FreeType(&ft_lib) != 0 {
		panic("FT_Init_FreeType failed")
	}
	defer C.FT_Done_FreeType(ft_lib)

	font_path_str := C.CString("/usr/share/fonts/opentype/firacode/FiraCode-Regular.otf")
	defer C.free(unsafe.Pointer(font_path_str))
	
	var ft_face C.FT_Face
	if err := C.FT_New_Face(ft_lib,
		font_path_str,
		0, &ft_face); err != 0 {
		panic(fmt.Sprintf("FT_New_Face failed:", err))
	}
	defer C.FT_Done_Face(ft_face)


	hb_buffer_ptr := C.hb_buffer_create()
	if hb_buffer_ptr == nil {
		panic("hb_buffer_create failed")
	}
	defer C.hb_buffer_destroy(hb_buffer_ptr)


	C.hb_buffer_set_direction(hb_buffer_ptr, C.HB_DIRECTION_LTR)
	C.hb_buffer_set_script(hb_buffer_ptr, C.HB_SCRIPT_LATIN)
	
	str := "😬 this is => my text"
	c_str := C.CString(str)
	defer C.free(unsafe.Pointer(c_str))
	c_len := C.int(len(str))

	C.hb_buffer_add_utf8(hb_buffer_ptr, c_str, c_len, 0, c_len)
	
	lang_str := C.CString("en")
	defer C.free(unsafe.Pointer(lang_str))
	C.hb_buffer_set_language(hb_buffer_ptr, C.hb_language_from_string(lang_str, -1))

	hb_font_ptr := C.hb_ft_font_create_referenced(ft_face)
	if hb_font_ptr == nil {
		panic("hb_ft_font_create failed")
	}
	defer C.hb_font_destroy(hb_font_ptr)
	C.hb_ft_font_set_funcs(hb_font_ptr)
	//font_load_flags := C.hb_ft_font_get_load_flags(hb_font_ptr)


	display_index := C.SDL_GetWindowDisplayIndex(win_ptr)
	var hdpi, vdpi C.float
	C.SDL_GetDisplayDPI(display_index, nil, &hdpi, &vdpi)
	fmt.Println("HDPI:", hdpi, "VDPI:", vdpi)
	
	C.FT_Set_Char_Size(ft_face,
		0, 24*64, // char width, height in 1/64th points
		C.uint(hdpi), C.uint(vdpi))
	//C.FT_Set_Pixel_Sizes(ft_face, 70, 70)
	C.hb_ft_font_changed(hb_font_ptr)
	

	C.hb_shape(hb_font_ptr, hb_buffer_ptr, nil, 0)
	var glyph_count C.uint
	_glyph_info := C.hb_buffer_get_glyph_infos(hb_buffer_ptr, &glyph_count)
	_glyph_position := C.hb_buffer_get_glyph_positions(hb_buffer_ptr, &glyph_count)
	glyph_info_arr := (**C.hb_glyph_info_t)(unsafe.Pointer(_glyph_info))
	glyph_position_arr := (**C.hb_glyph_position_t)(unsafe.Pointer(_glyph_position))


	var cursor_x, cursor_y = 100, 100
	for i, gc := 0, int(glyph_count); i < gc; i++ {
		gi := (*C.hb_glyph_info_t)(unsafe.Pointer(
			uintptr(unsafe.Pointer(glyph_info_arr)) +
			uintptr(i)*unsafe.Sizeof(*glyph_info_arr),
		))
		gp := (*C.hb_glyph_position_t)(unsafe.Pointer(
			uintptr(unsafe.Pointer(glyph_position_arr)) +
			uintptr(i)*unsafe.Sizeof(*glyph_position_arr),
		))
		
		glyph_id := gi.codepoint
		//fmt.Println(i, ":", glyph_id)
		x_offset := float32(gp.x_offset) / 64.0
		y_offset := float32(gp.y_offset) / 64.0
		x_advance := float32(gp.x_advance) / 64.0
		y_advance := float32(gp.y_advance) / 64.0
		x := cursor_x + int(x_offset)
		y := cursor_y + int(y_offset)

		if err := C.FT_Load_Glyph(ft_face, glyph_id, C.FT_LOAD_DEFAULT | C.FT_LOAD_COLOR); err != 0 {
			panic(fmt.Sprintf("FT_Load_Glyph failed:", err))
		}
		glyph_slot := ft_face.glyph
		C.FT_Render_Glyph(glyph_slot, C.FT_RENDER_MODE_NORMAL)
		glyph_bitmap := glyph_slot.bitmap
		glyph_surf_ptr := C.SDL_CreateRGBSurfaceFrom(
			unsafe.Pointer(glyph_bitmap.buffer),
			//C.int(glyph_bitmap.width),
			//C.int(glyph_bitmap.rows),
			C.int(glyph_slot.metrics.width),
			C.int(glyph_slot.metrics.height),
			32, C.int(glyph_bitmap.pitch),
			C.uint(0x000000ff),
			C.uint(0x0000ff00),
			C.uint(0x00ff0000),
			C.uint(0xff000000))
		if glyph_surf_ptr == nil {
			sdl_panic("SDL_CreateRGBSurfaceFrom(glyph_bitmap)", C.SDL_GetError())
		}
		C.SDL_SetSurfaceBlendMode(glyph_surf_ptr, C.SDL_BLENDMODE_BLEND)

		
		// var colors [256]C.SDL_Color
		// for i = 0; i < 256; i++ {
			// col := colors[i]
			// c_i := C.uchar(i)
			// col.r, col.b, col.g = c_i, c_i, c_i
		// }
		// C.SDL_SetPaletteColors(glyph_surf_ptr.format.palette, &colors[0], 0, 256)

		sdl_glyph_dest_rect := C.SDL_Rect {
			C.int(x), C.int(y),
			C.int(glyph_bitmap.width), C.int(glyph_bitmap.rows) }
			//C.int(glyph_slot.metrics.width), C.int(glyph_slot.metrics.height) }
			//0, 0}
		fmt.Println(sdl_glyph_dest_rect)
		C.SDL_FillRect(win_surf_ptr, nil, C.SDL_MapRGB(win_surf_ptr.format, 100, 0, 100))
		C.SDL_BlitSurface(glyph_surf_ptr, nil, win_surf_ptr, &sdl_glyph_dest_rect)

		C.SDL_FreeSurface(glyph_surf_ptr)

		cursor_x += int(x_advance)
		cursor_y += int(y_advance)
	}
	
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
