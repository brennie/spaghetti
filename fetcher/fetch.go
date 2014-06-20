// spaghetti: Applying Hierarchical Parallel Genetic Algorithms to solve the
// University Timetabling Problem.
// Copyright (C) 2014  Barret Rennie
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// The fetcher package for fetching instances from the web.
package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	nInstances = 24
	defaultDir = "instances"
	baseFile   = "comp-2007-2-%d.tim"
	baseUrl    = "http://www.cs.qub.ac.uk/itc2007/postenrolcourse/initialdatasets/comp-2007-2-%d.tim"
)

// Fetch the instances and store them in the given directory, or defaultDir if
// that is nil.
func Fetch(directory interface{}) {
	dir := defaultDir

	if directory != nil {
		dir = directory.(string)
	}

	err := os.Mkdir(dir, 0700)
	if err == nil || os.IsExist(err) {
		if err := os.Chdir(dir); err != nil {
			log.Fatal(err)
		}

		for i := 1; i <= nInstances; i++ {
			filename := fmt.Sprintf(baseFile, i)
			url := fmt.Sprintf(baseUrl, i)

			file, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}

			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}

			if resp.StatusCode != 200 {
				log.Fatalf("Got HTTP %s; expected HTTP 200 OK for %s", resp.StatusCode, url)
			}

			io.Copy(file, resp.Body)

			resp.Body.Close()
			file.Close()

			log.Printf("Downloaded %s%c%s instance (%d bytes)\n", dir, os.PathSeparator, filename, resp.ContentLength)
		}
	} else {
		log.Fatal(err)
	}
}
