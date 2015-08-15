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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	sessionServer = "https://sessionserver.mojang.com/session/minecraft/profile/"
)

var (
	imageCache = map[string]*imageEntry{}
	cacheLock  sync.RWMutex
)

type imageEntry struct {
	LastAccess atomic.Value
	Owner      string
	UUID       string
	Data       []byte
	CapeData   []byte
	wait       sync.WaitGroup
	Created    time.Time

	staleTime time.Duration
}

func init() {
	go imageGCWatcher()
}

func getOrCreateEntry(id string, create func(entry *imageEntry)) *imageEntry {
	cacheLock.RLock()
	entry, ok := imageCache[id]
	cacheLock.RUnlock()
	if !ok {
		cacheLock.Lock()
		entry, ok = imageCache[id]
		// Might have been created between locks
		if !ok {
			entry = &imageEntry{
				UUID:    id,
				Created: time.Now(),
			}
			entry.LastAccess.Store(time.Now())
			imageCache[id] = entry
			entry.wait.Add(1)
			cacheLock.Unlock()
			entry.staleTime = Config.IconStaleTime
			create(entry)
			entry.wait.Done()
		} else {
			cacheLock.Unlock()
		}
	}
	entry.wait.Wait() // Make sure it exists
	entry.LastAccess.Store(time.Now())
	return entry
}

func getSkinForID(id string) *imageEntry {
	if id[12] != '4' {
		return &imageEntry{} // Offline mode
	}
	entry := getOrCreateEntry(id, func(entry *imageEntry) {
		if err := getEntry(entry); err != nil {
			// TODO Default skin
			entry.Data = nil
		}
		entry.staleTime = Config.ImageStaleTime
	})
	return entry
}

func getEntry(entry *imageEntry) error {
	resp, err := http.Get(fmt.Sprintf("%s%s", sessionServer, entry.UUID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var reply mojangSessionReply
	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return err
	}
	var texInfo string
	for _, p := range reply.Properties {
		if p.Name == "textures" {
			texInfo = p.Value
			break
		}
	}
	if texInfo == "" {
		return errors.New("Missing textures")
	}
	data, err := base64.StdEncoding.DecodeString(texInfo)
	if err != nil {
		return err
	}
	var tex mojangTextureInfo
	err = json.Unmarshal(data, &tex)
	if err != nil {
		return err
	}
	if tex.Textures.Skin.URL != "" {
		resp, err := http.Get(tex.Textures.Skin.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		entry.Data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}
	if tex.Textures.Cape.URL != "" {
		resp, err := http.Get(tex.Textures.Cape.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		entry.CapeData, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

type mojangSessionReply struct {
	ID         string
	Name       string
	Properties []struct {
		Name  string
		Value string
	}
}

type mojangTextureInfo struct {
	Textures struct {
		Skin struct{ URL string }
		Cape struct{ URL string }
	}
}

func imageGCWatcher() {
	for {
		time.Sleep(Config.ImageGCInterval)
		imageGC()
	}
}

func imageGC() {
	log.Println("Cleaning up old images")
	cacheLock.Lock()
	defer cacheLock.Unlock()
	now := time.Now()
	for k, v := range imageCache {
		lastAccess := v.LastAccess.Load().(time.Time)
		if now.Sub(lastAccess) > v.staleTime || now.Sub(v.Created) > Config.RefetchTime {
			delete(imageCache, k)
			log.Println("Removed ", k)
		}
	}
}
