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
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Try and load the config
	err := loadConfig()
	if os.IsNotExist(err) {
		// First time starting
		log.Println("Config not found, creating")
		saveConfig()
		return
	}
	if err != nil {
		panic(err)
	}
	saveConfig()

	r := mux.NewRouter()

	// Basic types
	r.HandleFunc("/icons/raw/{uuid:[a-f0-9]{32}}", rawSkin)

	r.HandleFunc("/icons/head/{uuid:[a-f0-9]{32}}", basicIcon)
	r.HandleFunc("/icons/head/{uuid:[a-f0-9]{32}}/{hat:hat}", basicIcon)
	r.HandleFunc("/icons/head/{uuid:[a-f0-9]{32}}/{size:[0-9]+}", basicIcon)
	r.HandleFunc("/icons/head/{uuid:[a-f0-9]{32}}/{hat:hat}/{size:[0-9]+}", basicIcon)

	log.Printf("Starting mcicon at http://%s:%d", Config.Host, Config.Port)
	log.Println(
		http.ListenAndServe(fmt.Sprintf("%s:%d", Config.Host, Config.Port), r),
	)
}
