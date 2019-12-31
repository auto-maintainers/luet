// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

package client

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	. "github.com/mudler/luet/pkg/logger"

	"github.com/mudler/luet/pkg/compiler"
	"github.com/mudler/luet/pkg/helpers"

	"github.com/cavaliercoder/grab"
)

type HttpClient struct {
	RepoData RepoData
}

func NewHttpClient(r RepoData) *HttpClient {
	return &HttpClient{RepoData: r}
}

func (c *HttpClient) DownloadArtifact(artifact compiler.Artifact) (compiler.Artifact, error) {
	var file *os.File = nil
	var u *url.URL = nil

	artifactName := path.Base(artifact.GetPath())
	ok := false

	temp, err := ioutil.TempDir(os.TempDir(), "tree")
	if err != nil {
		return nil, err
	}

	for _, uri := range c.RepoData.Urls {
		Info("Downloading artifact", artifactName, "from", uri)

		file, err = ioutil.TempFile(temp, "HttpClient")
		if err != nil {
			continue
		}

		u, err = url.Parse(uri)
		if err != nil {
			continue
		}
		u.Path = path.Join(u.Path, artifactName)

		_, err = grab.Get(temp, u.String())
		if err != nil {
			continue
		}

		Debug("Copying file ", filepath.Join(temp, artifactName), "to", file.Name())
		err = helpers.CopyFile(filepath.Join(temp, artifactName), file.Name())
		if err != nil {
			continue
		}
		ok = true
		break
	}

	if !ok {
		return nil, err
	}

	newart := artifact
	newart.SetPath(file.Name())
	return newart, nil
}

func (c *HttpClient) DownloadFile(name string) (string, error) {
	var file *os.File = nil
	var u *url.URL = nil
	ok := false

	temp, err := ioutil.TempDir(os.TempDir(), "tree")
	if err != nil {
		return "", err
	}

	for _, uri := range c.RepoData.Urls {

		file, err = ioutil.TempFile(os.TempDir(), "HttpClient")
		if err != nil {
			continue
		}
		//defer os.Remove(file.Name())
		u, err = url.Parse(uri)
		if err != nil {
			continue
		}
		u.Path = path.Join(u.Path, name)

		Info("Downloading", u.String())

		_, err = grab.Get(temp, u.String())
		if err != nil {
			continue
		}

		err = helpers.CopyFile(filepath.Join(temp, name), file.Name())
		if err != nil {
			continue
		}
		ok = true
		break
	}

	if !ok {
		return "", err
	}

	return file.Name(), err
}
