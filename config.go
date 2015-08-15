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
	"encoding/json"
	"os"
	"time"
)

var Config = ConfigType{
	Host:            "localhost",
	Port:            9445,
	ImageGCInterval: time.Second * 20,
	ImageStaleTime:  time.Minute * 10,
	IconStaleTime:   time.Second * 20,
	RefetchTime:     time.Hour * 24,
	MaxSize:         512,
	MinSize:         8,
}

type ConfigType struct {
	Host            string
	Port            int
	ImageGCInterval time.Duration
	ImageStaleTime  time.Duration
	IconStaleTime   time.Duration
	RefetchTime     time.Duration

	MaxSize int
	MinSize int
}

func saveConfig() {
	f, err := os.Create("mcicon.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	data, err := json.MarshalIndent(&Config, "", "    ")
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
}

func loadConfig() error {
	f, err := os.Open("mcicon.json")
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		return err
	}
	return nil
}
