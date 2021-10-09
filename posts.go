package instag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	neutral "net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	media    Media
	Images          Images   `json:"image_versions2,omitempty"`
	OriginalWidth   int      `json:"original_width,omitempty"`
	OriginalHeight  int      `json:"original_height,omitempty"`
	ImportedTakenAt int64    `json:"imported_taken_at,omitempty"`
	Location        Location `json:"location,omitempty"`
	Lat             float64  `json:"lat,omitempty"`
	Lng             float64  `json:"lng,omitempty"`

}

func (insta *Instagram) UploadPhoto(photo io.Reader, photoCaption string, quality int, filterType int) (Item, error) {
	out := Item{}

	config, err := insta.postPhoto(photo, photoCaption, quality, filterType, false)
	if err != nil {
		return out, err
	}
	data, err := insta.prepareData(config)
	if err != nil {
		return out, err
	}

	body, err := insta.sendRequest(&reqOptions{
		Endpoint: "media/configure/?",
		Query:    generateSignature(data),
		IsPost:   true,
	})
	if err != nil {
		return out, err
	}
	var uploadResult struct {
		Media    Item   `json:"media"`
		UploadID string `json:"upload_id"`
		Status   string `json:"status"`
	}
	err = json.Unmarshal(body, &uploadResult)
	if err != nil {
		return out, err
	}

	if uploadResult.Status != "ok" {
		return out, fmt.Errorf("invalid status, result: %s", uploadResult.Status)
	}

	return uploadResult.Media, nil
}

func (insta *Instagram) postPhoto(photo io.Reader, photoCaption string, quality int, filterType int, isSidecar bool) (map[string]interface{}, error) {
	uploadID := time.Now().Unix()
	photoName := fmt.Sprintf("pending_media_%d.jpg", uploadID)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("upload_id", strconv.FormatInt(uploadID, 10))
	w.WriteField("_uuid", insta.uuid)
	w.WriteField("_csrftoken", insta.token)
	var compression = map[string]interface{}{
		"lib_name":    "jt",
		"lib_version": "1.3.0",
		"quality":     quality,
	}
	cBytes, _ := json.Marshal(compression)
	w.WriteField("image_compression", toString(cBytes))
	if isSidecar {
		w.WriteField("is_sidecar", toString(1))
	}
	fw, err := w.CreateFormFile("photo", photoName)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	rdr := io.TeeReader(photo, &buf)
	if _, err = io.Copy(fw, rdr); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", goInstaAPIUrl+"upload/photo/", &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-IG-Capabilities", "3Q4=")
	req.Header.Set("X-IG-Connection-Type", "WIFI")
	req.Header.Set("Cookie2", "$Version=1")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-type", w.FormDataContentType())
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", goInstaUserAgent)

	resp, err := insta.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code, result: %s", resp.Status)
	}
	var result struct {
		UploadID       string      `json:"upload_id"`
		XsharingNonces interface{} `json:"xsharing_nonces"`
		Status         string      `json:"status"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.Status != "ok" {
		return nil, fmt.Errorf("unknown error, status: %s", result.Status)
	}
	width, height, err := getImageDimensionFromReader(&buf)
	if err != nil {
		return nil, err
	}
	config := map[string]interface{}{
		"media_folder": "Instagram",
		"source_type":  4,
		"caption":      photoCaption,
		"upload_id":    strconv.FormatInt(uploadID, 10),
		"device":       goInstaDeviceSettings,
		"edits": map[string]interface{}{
			"crop_original_size": []int{width * 1.0, height * 1.0},
			"crop_center":        []float32{0.0, 0.0},
			"crop_zoom":          1.0,
			"filter_type":        filterType,
		},
		"extra": map[string]interface{}{
			"source_width":  width,
			"source_height": height,
		},
	}
	return config, nil
}
