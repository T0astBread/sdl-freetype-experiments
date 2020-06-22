package main

/*
#cgo LDFLAGS: -lSDL2 -lharfbuzz -lfreetype
#cgo pkg-config: sdl2 harfbuzz freetype2

#include <SDL.h>
#include <hb.h>
#include <hb-ft.h>
#include <freetype/freetype.h>

SDL_EventType sdl_event_type(SDL_Event* evt) {
	return evt->type;
}

int ft_has_color(FT_Face ft_face) {
	return FT_HAS_COLOR(ft_face);
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
	//win_surf_ptr := C.SDL_GetWindowSurface(win_ptr)

	renderer_ptr := C.SDL_CreateRenderer(win_ptr, -1, C.SDL_RENDERER_ACCELERATED | C.SDL_RENDERER_PRESENTVSYNC)
	if renderer_ptr == nil {
		sdl_panic("SDL_CreateRenderer", C.SDL_GetError())
	}
	defer C.SDL_DestroyRenderer(renderer_ptr)


	var ft_lib C.FT_Library
	if C.FT_Init_FreeType(&ft_lib) != 0 {
		panic("FT_Init_FreeType failed")
	}
	defer C.FT_Done_FreeType(ft_lib)

	//font_path_str := C.CString("/usr/share/fonts/opentype/firacode/FiraCode-Regular.otf")
	font_path_str := C.CString("/usr/share/fonts/truetype/noto/NotoColorEmoji.ttf")
	//font_path_str := C.CString("/usr/share/fonts/opentype/freefont/FreeSans.otf")
	//font_path_str := C.CString("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	//font_path_str := C.CString("/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf")
	defer C.free(unsafe.Pointer(font_path_str))
	
	var ft_face C.FT_Face
	if err := C.FT_New_Face(ft_lib,
		font_path_str,
		0, &ft_face); err != 0 {
		panic(fmt.Sprintf("FT_New_Face failed:", err))
	}
	defer C.FT_Done_Face(ft_face)

	fmt.Println("Has color:", ft_has_color(ft_face))


	hb_buffer_ptr := C.hb_buffer_create()
	if hb_buffer_ptr == nil {
		panic("hb_buffer_create failed")
	}
	defer C.hb_buffer_destroy(hb_buffer_ptr)


	C.hb_buffer_set_direction(hb_buffer_ptr, C.HB_DIRECTION_LTR)
	C.hb_buffer_set_script(hb_buffer_ptr, C.HB_SCRIPT_LATIN)
	//C.hb_buffer_set_script(hb_buffer_ptr, C.HB_SCRIPT_MANDAIC)
	//C.hb_buffer_set_direction(hb_buffer_ptr, C.HB_DIRECTION_RTL)
	//C.hb_buffer_set_script(hb_buffer_ptr, C.HB_SCRIPT_ARABIC)
	
	//str := "😬 this is => my text\nöèẁ²"
	//str := "هذه هي بعض النصوص العربي"
	//str := "這是一些中文"
	//str := "😬😬😬😬😬😬😬😬😬😬😬😬😬"
	str := "🏳️‍🌈🏳️‍🌈🥳🥳"
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
	//C.hb_ft_font_set_funcs(hb_font_ptr)
	//font_load_flags := C.hb_ft_font_get_load_flags(hb_font_ptr)


	display_index := C.SDL_GetWindowDisplayIndex(win_ptr)
	var hdpi, vdpi C.float
	C.SDL_GetDisplayDPI(display_index, nil, &hdpi, &vdpi)
	fmt.Println("HDPI:", hdpi, "VDPI:", vdpi)

	if ft_has_color(ft_face) {
		C.FT_Select_Size(ft_face, 0)
	} else {
		if err := C.FT_Set_Char_Size(ft_face,
			0, 24*64, // char width, height in 1/64th points
			C.uint(hdpi), C.uint(vdpi));
		err != 0 {
			panic(fmt.Sprintf("FT_Set_Char_Size failed:", err))
		}
		//C.FT_Set_Pixel_Sizes(ft_face, 128, 0)
	}
	C.hb_ft_font_changed(hb_font_ptr)
	

	C.hb_shape(hb_font_ptr, hb_buffer_ptr, nil, 0)
	var glyph_count C.uint
	_glyph_info := C.hb_buffer_get_glyph_infos(hb_buffer_ptr, &glyph_count)
	_glyph_position := C.hb_buffer_get_glyph_positions(hb_buffer_ptr, &glyph_count)
	//glyph_info_arr := (**C.hb_glyph_info_t)(unsafe.Pointer(_glyph_info))
	//glyph_position_arr := (**C.hb_glyph_position_t)(unsafe.Pointer(_glyph_position))
	glyph_position_arr := (*[1 << 28]C.hb_glyph_position_t)(unsafe.Pointer(_glyph_position))[:glyph_count:glyph_count]
	glyph_info_arr := (*[1 << 28]C.hb_glyph_info_t)(unsafe.Pointer(_glyph_info))[:glyph_count:glyph_count]

	C.SDL_SetRenderDrawColor(renderer_ptr, C.uchar(100), C.uchar(0), C.uchar(100), C.uchar(255))
	C.SDL_RenderClear(renderer_ptr)

	var cursor_x, cursor_y = 100, 100
	for i, gc := 0, int(glyph_count); i < gc; i++ {
		// gi := (*C.hb_glyph_info_t)(unsafe.Pointer(
			// uintptr(unsafe.Pointer(glyph_info_arr)) +
			// uintptr(i)*unsafe.Sizeof(*glyph_info_arr),
		// ))
		// gp := (*C.hb_glyph_position_t)(unsafe.Pointer(
			// uintptr(unsafe.Pointer(glyph_position_arr)) +
			// uintptr(i)*unsafe.Sizeof(*glyph_position_arr),
		// ))
		gi := glyph_info_arr[i]
		gp := glyph_position_arr[i]
		
		glyph_id := gi.codepoint
		fmt.Println(i, ":", glyph_id)
		x_offset := float32(gp.x_offset) / 64.0
		y_offset := float32(gp.y_offset) / 64.0
		//y_offset := 0
		x_advance := float32(gp.x_advance) / 64.0
		y_advance := float32(gp.y_advance) / 64.0
		//y_advance := 0

		var load_flags C.int = C.FT_LOAD_DEFAULT
		if ft_has_color(ft_face) {
			load_flags |= C.FT_LOAD_COLOR
		} else {
			load_flags |= C.FT_LOAD_RENDER
		}
		if err := C.FT_Load_Glyph(ft_face, glyph_id, load_flags); err != 0 {
			//panic(fmt.Sprintf("FT_Load_Glyph failed:", err))
			fmt.Println(i, "FT_Load_Glyph failed:", err)
			continue
		}
		glyph_slot := ft_face.glyph
		//C.FT_Render_Glyph(glyph_slot, C.FT_RENDER_MODE_NORMAL)
		glyph_bitmap := glyph_slot.bitmap
		//glyph_texture := SDLTextureFromFTBitmap(renderer_ptr, glyph_bitmap, C.SDL_Color { 200, 200, 200, 255 })
		glyph_texture := SDLTextureFromFTColorBitmap(renderer_ptr, glyph_bitmap)
		if glyph_texture == nil {
			println("nil")
			//continue
		}

		var glyph_w, glyph_h C.int = C.int(glyph_bitmap.width), C.int(glyph_bitmap.rows)
		//C.SDL_QueryTexture(glyph_texture, nil, nil, &glyph_w, &glyph_h)

		x := cursor_x + int(x_offset) + int(glyph_slot.bitmap_left)
		//y := cursor_y + int(y_offset)
		y:= C.int(cursor_y) + C.int(y_offset) - glyph_slot.bitmap_top
		
		//fmt.Println("Glyph dimen:", glyph_w, glyph_h)
		//fmt.Println("Glyph offset:", x, y)
		glyph_rect := C.SDL_Rect { C.int(x), C.int(y), glyph_w, glyph_h }
		
		C.SDL_SetTextureBlendMode(glyph_texture, C.SDL_BLENDMODE_BLEND)
		// C.SDL_SetRenderDrawColor(renderer_ptr, C.uchar(50), C.uchar(0), C.uchar(25), C.uchar(255))
		// C.SDL_RenderFillRect(renderer_ptr, &glyph_rect);
		C.SDL_RenderCopy(renderer_ptr, glyph_texture, nil, &glyph_rect)

		C.SDL_DestroyTexture(glyph_texture)

		cursor_x_adv := int(x_advance)// + int(glyph_w) + int(glyph_slot.bitmap_left)
		cursor_x += cursor_x_adv
		cursor_y -= int(y_advance)

		//fmt.Println("Cursor +=", cursor_x_adv)
	}

	C.SDL_RenderPresent(renderer_ptr)

	running := true
	for running == true {
		var event C.SDL_Event
		for C.SDL_PollEvent(&event) != 0 {
			switch C.sdl_event_type(&event) {
				case C.SDL_QUIT:
					running = false
			}
		}

		//C.SDL_UpdateWindowSurface(win_ptr)
		C.SDL_Delay(16)
	}
}

func SDLTextureFromFTColorBitmap(
	renderer *C.SDL_Renderer,
	bitmap C.FT_Bitmap,
) *C.SDL_Texture {
	w, h := int(bitmap.width), int(bitmap.rows)
	_w, _h := C.int(w), C.int(h)

	if w == 0 || h == 0 {
		return nil
	}
	
	texture := C.SDL_CreateTexture(renderer,
		C.SDL_PIXELFORMAT_RGBA8888,
		C.SDL_TEXTUREACCESS_STREAMING,
		_w, _h)
		
	if texture == nil {
		sdl_panic("SDL_CreateTexture", C.SDL_GetError())
	}

	var buffer unsafe.Pointer
	var pitch C.int
	C.SDL_LockTexture(texture, nil, &buffer, &pitch)
	defer C.SDL_UnlockTexture(texture)

	pixelFormat := C.SDL_AllocFormat(C.SDL_PIXELFORMAT_RGBA8888)
	if pixelFormat == nil {
		sdl_panic("SDL_AllocFormat", C.SDL_GetError())
	}
	defer C.SDL_FreeFormat(pixelFormat)

	numBytes := _w * _h
	source := C.GoBytes(unsafe.Pointer(bitmap.buffer), numBytes * 4)
	//target := C.GoBytes(buffer, numBytes)
	target := (*[1 << 28]C.uint)(unsafe.Pointer(buffer))[:numBytes:numBytes]

	// for y := 0; y < h; y++ {
		// for x := 0; x < w; x+=4 {
			// idx := y * w + x
			// b := source[idx]
			// g := source[idx+1]
			// r := source[idx+2]
			// a := source[idx+3]
			// pixelValue := C.SDL_MapRGBA(pixelFormat, C.uchar(r), C.uchar(g),C.uchar(b), C.uchar(a))
			// //target[idx] = byte(pixelValue)
			// target[idx/4] = pixelValue
		// }
	// }
	_numBytes := int(numBytes)
	for i := 0; i < _numBytes; i++ {
		idx := i * 4
		
		b := source[idx]
		g := source[idx+1]
		r := source[idx+2]
		a := source[idx+3]
		pixelValue := C.SDL_MapRGBA(pixelFormat, C.uchar(r), C.uchar(g),C.uchar(b), C.uchar(a))
		//target[idx] = byte(pixelValue)
		target[i] = pixelValue
	}

	return texture
}

func SDLTextureFromFTBitmap(
	renderer *C.SDL_Renderer,
	bitmap C.FT_Bitmap,
	color C.SDL_Color,
) *C.SDL_Texture {
	w, h := int(bitmap.width), int(bitmap.rows)
	_w, _h := C.int(w), C.int(h)

	if w == 0 || h == 0 {
		return nil
	}
	
	texture := C.SDL_CreateTexture(renderer,
		C.SDL_PIXELFORMAT_RGBA8888,
		C.SDL_TEXTUREACCESS_STREAMING,
		_w, _h)
		
	if texture == nil {
		sdl_panic("SDL_CreateTexture", C.SDL_GetError())
	}

	var buffer unsafe.Pointer
	var pitch C.int
	C.SDL_LockTexture(texture, nil, &buffer, &pitch)
	defer C.SDL_UnlockTexture(texture)

	pixelFormat := C.SDL_AllocFormat(C.SDL_PIXELFORMAT_RGBA8888)
	if pixelFormat == nil {
		sdl_panic("SDL_AllocFormat", C.SDL_GetError())
	}
	defer C.SDL_FreeFormat(pixelFormat)

	numBytes := _w * _h
	source := C.GoBytes(unsafe.Pointer(bitmap.buffer), numBytes)
	//target := C.GoBytes(buffer, numBytes)
	target := (*[1 << 28]C.uint)(unsafe.Pointer(buffer))[:numBytes:numBytes]

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y * w + x
			alpha := source[idx]
			pixelValue := C.SDL_MapRGBA(pixelFormat, color.r, color.g, color.b, C.uchar(alpha))
			//target[idx] = byte(pixelValue)
			target[idx] = pixelValue
		}
	}

	return texture
}

func ft_has_color(ft_face C.FT_Face) bool {
	return C.ft_has_color(ft_face) != 0
}
