// Copyright (C) 2023 Patrice Congo <@congop>
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

package stubrepo

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type PullImgSpecData struct {
	Tag       saya.Reference
	Hash      string
	OsVariant string
	Platform  saya.Platform
	RepoType  string
	ImgType   string
}

type DummyImg struct {
	img       []byte
	metaData  *saya.ImageTagMetaData
	meta      []byte // cached value of serialized metaData
	withMeta  bool   // true if meta as byte can be served if metaData available
	BasicAuth *saya.AuthHttpBasic
	UrlPath   string
}

func PersistImgMeta(target io.Writer, imgMeta *saya.ImageTagMetaData) error {
	encoder := yaml.NewEncoder(target)
	defer encoder.Close()

	if err := encoder.Encode(imgMeta); err != nil {
		return errors.Wrapf(err, "fail to encode image meta %v", imgMeta)
	}

	return nil
}

func (img *DummyImg) MetaDataAsBytes() ([]byte, error) {
	if !img.withMeta {
		// meta should not be served
		return nil, nil
	}
	if img.metaData == nil {
		return img.meta, nil
	}
	metaDataBuf := bytes.Buffer{}
	if err := PersistImgMeta(&metaDataBuf, img.metaData); err != nil {
		return nil, err
	}
	img.meta = metaDataBuf.Bytes()
	return img.meta, nil
}

func NewDummyImg(
	img []byte, withMeta bool,
	pullSpecData PullImgSpecData,
	basicAuth *saya.AuthHttpBasic,
) (*DummyImg, error) {
	if len(img) > 128 {
		return nil, errors.Errorf("NewDummyImg -- image too big: max-byte-size=128, actual-size=%d", len(img))
	}

	sha256, err := DigestSha256(bytes.NewReader(img))
	if err != nil {
		return nil, err
	}

	urlPathSegments := ImageRepoDirRelUrlSegments(&pullSpecData.Platform, pullSpecData.Tag.Name, pullSpecData.Tag.Version)
	imageFileBasename, err := ImgFileName(pullSpecData.ImgType)
	if err != nil {
		return nil, err
	}
	urlPath := strings.Join(append(urlPathSegments, imageFileBasename), "/")

	metaData := saya.ImageTagMetaData{
		Name:      pullSpecData.Tag.Name,
		Version:   pullSpecData.Tag.Version,
		Sha256:    sha256,
		Type:      pullSpecData.ImgType,
		Platform:  saya.PlatformSw{Platform: pullSpecData.Platform, OsVariant: pullSpecData.OsVariant},
		CreatedAt: time.Now(),
		SrcType:   pullSpecData.RepoType,
	}

	dummyImg := DummyImg{
		img:       img,
		metaData:  &metaData,
		meta:      nil,
		withMeta:  withMeta,
		BasicAuth: basicAuth,
		UrlPath:   urlPath,
	}

	return &dummyImg, nil
}

type HttpRepo struct {
	ramImgStore map[string]*DummyImg
	ip          string
	port        uint16
	httpServer  *http.Server
}

func NewDummyHttpRepo() *HttpRepo {
	return &HttpRepo{ramImgStore: map[string]*DummyImg{}}
}

func (repo *HttpRepo) RegisterDummyImg(img *DummyImg) error {
	if img == nil {
		return errors.Errorf("DummyHttpRepo.RegisterDummyImg -- img must nor be nil")
	}
	repo.ramImgStore[img.UrlPath] = img
	repo.ramImgStore[img.UrlPath+".meta"] = img
	return nil
}

func (repo *HttpRepo) AsRepos() *saya.Repos {
	return &saya.Repos{
		Http: &saya.HttpRepo{RepoUrl: fmt.Sprintf("http://%s:%d", repo.ip, repo.port), BasePath: "repo"},
	}
}

func (repo *HttpRepo) RepoUrl(relPath string) string {
	relPathNormalized := strings.TrimLeft(strings.TrimSpace(relPath), "/")
	return fmt.Sprintf("http://%s:%d/repo/%s", repo.ip, repo.port, relPathNormalized)
}

func (repo *HttpRepo) Close() error {
	if repo.httpServer == nil {
		return nil
	}
	srv := repo.httpServer
	repo.httpServer = nil

	return srv.Close()
}

func (repo *HttpRepo) Start() error {
	engine := gin.New()
	// engine.Handle("GET", "/repo", repo.GetData)
	engine.NoRoute(repo.NoRoot)

	addrStr := "localhost:0"
	ln, err := net.Listen("tcp", addrStr)
	if err != nil {
		return errors.Wrapf(
			err, "DummyHttpRepo.Start -- fail to start listening at: addr=%s err=%v",
			addrStr, err)
	}
	addr := ln.Addr()
	srv := &http.Server{
		Addr:    addr.String(),
		Handler: engine,
	}
	repo.httpServer = srv
	go func() {
		var err error
		log.Debugf(StubLogCtx(), "DummyHttpRepo.Start -- serving at: addr=%v", addr)
		if err = srv.Serve(ln); err != nil {
			if err != http.ErrServerClosed {
				log.Infof(StubLogCtx(), "DummyHttpRepo.Start -- server closed: addr=%v", addr)
				return
			}
		}
		log.Infof(StubLogCtx(), "DummyHttpRepo.Start --  server stop serving: addr=%v, err=%+v", addr, err)
	}()

	lIp, lPort, err := net.SplitHostPort(addr.String()) //revive:disable-line:var-naming
	if err != nil {
		return errors.Wrapf(err, "DummyHttpRepo.Start -- fail to split host-port: addr=%v", addr)
	}
	lPortUint64, _ := strconv.ParseUint(lPort, 10, 16)

	repo.ip = lIp
	repo.port = uint16(lPortUint64)

	return nil
}
func (repo *HttpRepo) NoRoot(ginCtx *gin.Context) {
	method := ginCtx.Request.Method
	resPath := ginCtx.Request.URL.Path
	log.Infof(StubLogCtx(), "DummyHttpRepo.NoRoot -- request: method=%s resPath=%s available=%v", method, resPath, maps.Keys(repo.ramImgStore))

	switch method {
	case "HEAD":
		ginCtx.Status(http.StatusNotImplemented)
	case "GET":
		repo.GetData(ginCtx)
	default:
		log.Infof(StubLogCtx(), "DummyHttpRepo.NoRoot -- not found: method=%s resPath=%s available=%v", method, resPath, maps.Keys(repo.ramImgStore))
		ginCtx.Status(http.StatusNotFound)
	}
}

func (repo *HttpRepo) RepoContent() map[string]string {
	entries := map[string]string{}
	for k, v := range repo.ramImgStore {
		vStr := ""
		if meta := v.metaData; meta != nil {
			vStr = strings.Join([]string{meta.Platform.PlatformStr(), meta.Name, meta.Version, meta.Type}, ":")
		} else {
			vStr = v.UrlPath
		}
		entries[k] = vStr
	}
	return entries
}

func (repo *HttpRepo) GetData(ginCtx *gin.Context) {
	resPathHttp := ginCtx.Request.URL.Path
	log.Infof(StubLogCtx(), "DummyHttpRepo.GetData -- request: resPathHttp=%s available=%v", resPathHttp, maps.Keys(repo.ramImgStore))
	resPath := strings.TrimPrefix(resPathHttp, "/repo/")
	dummyImg, avail := repo.ramImgStore[resPath]
	if dummyImg == nil || !avail {

		log.Infof(StubLogCtx(), "DummyHttpRepo.GetData -- not found: resPath=%s available=%v", resPath, maps.Keys(repo.ramImgStore))
		ginCtx.Data(http.StatusNotFound, "", []byte("no dummy img found:"+resPath))
		return
	}
	if dummyImg.BasicAuth != nil {
		u, p, ok := ginCtx.Request.BasicAuth()
		if !ok || u != dummyImg.BasicAuth.Username || p != dummyImg.BasicAuth.Pwd {
			// Todo add WWW-Authenticate: Basic realm="User Visible Realm"
			// @see https://en.wikipedia.org/wiki/Basic_access_authentication
			ginCtx.Status(http.StatusForbidden)
			return
		}
	}
	switch {
	case strings.HasSuffix(resPath, dummyImg.UrlPath):
		ginCtx.Data(http.StatusOK, "", dummyImg.img)
	case strings.HasSuffix(resPath, ".meta") && dummyImg.metaData != nil:
		metaByte, err := dummyImg.MetaDataAsBytes()
		if err != nil {
			log.Errorf(StubLogCtx(), "DummyHttpRepo.GetData -- fail to make meta bytes: path=%s, err=%+v", resPath, err)
		}
		ginCtx.Data(http.StatusOK, "", metaByte)
	default:
		ginCtx.Status(http.StatusNotFound)
	}

}
