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

	"golang.org/x/image/draw"

	"github.com/gorilla/mux"
)

func rawSkin(rw http.ResponseWriter, rq *http.Request) {
	v := mux.Vars(rq)
	skin := getSkinForID(v["uuid"])
	rw.Write(skin.Data)
}

func basicIcon(rw http.ResponseWriter, rq *http.Request) {
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

	id := fmt.Sprintf("head:%d:%s:%s", size, hat, v["uuid"])
	rw.Write(getOrCreateEntry(id, func(entry *imageEntry) {
		skin := getSkinForID(v["uuid"])
		img, err := png.Decode(bytes.NewReader(skin.Data))
		if err != nil {
			return
		}
		out := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.NearestNeighbor.Scale(
			out,
			image.Rect(0, 0, size, size),
			img,
			image.Rect(8, 8, 8+8, 8+8),
			draw.Over,
			nil,
		)
		if hat == "hat" {
			draw.NearestNeighbor.Scale(
				out,
				image.Rect(0, 0, size, size),
				img,
				image.Rect(32+8, 8, 32+8+8, 8+8),
				draw.Over,
				nil,
			)
		}

		var buf bytes.Buffer
		png.Encode(&buf, out)
		entry.Data = buf.Bytes()
	}).Data)
}
