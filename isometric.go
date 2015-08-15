// Copyright 2015 Matthew Collins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func isoHead(rw http.ResponseWriter, rq *http.Request) {
	v := mux.Vars(rq)
	size, err := strconv.Atoi(v["size"])
	if err != nil {
		size = 64
	}
	if size > Config.MaxSize {
		size = Config.MaxSize
	} else if size < Config.MinSize {
		size = Config.MinSize
	}
	hat := v["hat"]
	if hat == "" {
		hat = "nohat"
	}

	id := fmt.Sprintf("head_iso:%d:%s:%s", size, hat, v["uuid"])
	rw.Write(getOrCreateEntry(id, func(entry *imageEntry) {
		skin := getSkinForID(v["uuid"])
		img, err := png.Decode(bytes.NewReader(skin.Data))
		if err != nil {
			return
		}
		top := image.Rect(8, 0, 16, 8)
		left := image.Rect(0, 8, 8, 16)
		right := image.Rect(8, 8, 16, 16)

		out := image.NewNRGBA(image.Rect(0, 0, size, size))

		fo := 0
		fs := size
		if hat == "hat" {
			fo += size / 32
			fs -= fo * 2
		}

		drawIsometricCube(out, fo, fo, fs, fs, img, top, left, right)

		top = image.Rect(32+8, 0, 32+16, 8)
		left = image.Rect(32+0, 8, 32+8, 16)
		right = image.Rect(32+8, 8, 32+16, 16)
		drawIsometricCube(out, 0, 0, size, size, img, top, left, right)

		var buf bytes.Buffer
		png.Encode(&buf, out)
		entry.Data = buf.Bytes()
	}).Data)
}

func drawIsometricCube(out *image.NRGBA, x, y, w, h int, src image.Image, top, left, right image.Rectangle) {
	for tx := 0; tx < w/2; tx++ {
		for ty := 0; ty < h/2; ty++ {
			col := src.At(
				left.Min.X+int((float64(tx)/float64(w/2))*float64(left.Dx())),
				left.Min.Y+int((float64(ty)/float64(h/2))*float64(left.Dy())),
			)
			if _, _, _, a := col.RGBA(); a == 0xFFFF {
				out.Set(
					x+tx,
					(h/4)+y+ty+int(float64(tx)*0.5),
					col,
				)
			}

			col = src.At(
				right.Min.X+int((float64(tx)/float64(w/2))*float64(right.Dx())),
				right.Min.Y+int((float64(ty)/float64(h/2))*float64(right.Dy())),
			)
			if _, _, _, a := col.RGBA(); a == 0xFFFF {
				out.Set(
					x+tx+(w/2),
					(h/4)+y+ty+int(float64((w/2)-tx)*0.5),
					col,
				)
			}
		}
	}

	for ttx := -1; ttx < (w/2)+1; ttx++ {
		for tty := -1; tty < (h/2)+1; tty++ {
			tx, ty := clamp(ttx, 0, w/2), clamp(tty, 0, h/2)
			col := src.At(
				top.Min.X+int((float64(tx)/float64(w/2))*float64(top.Dx())),
				top.Min.Y+int((float64(ty)/float64(h/2))*float64(top.Dy())),
			)
			if _, _, _, a := col.RGBA(); a == 0xFFFF {
				out.Set(
					x+1+ttx+tty,
					(h/4)+y+int(float64(ttx)*0.5-0.75-float64(tty)*0.5),
					col,
				)
			}
		}
	}
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
