package attach

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/yuansfer/log"

	"github.com/yuansfer/asana/util"
)

type Request struct {
	Body     io.Reader
	TaskID   string
	Name     string
	Path     string
	PAToken  string
	TicketID string
}

type Data struct {
	Response Response `json:"data"`
}

type Response struct {
	ID           string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	// ResourceSubtype - asana, dropbox, gdrive, onedrive, box, and external
	ResourceSubtype string `json:"resource_subtype"`
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (r Request) CreateInEncryption(download func(src, tmp string) error) (*Response, error) {
	return r.create(download)
}

// Create - upload attachments
func (r Request) Create() (resp *Response, err error) {
	return r.create(download)
}

func (r Request) create(f func(src, tmp string) error) (resp *Response, err error) {
	var (
		tempPath string
		data     struct {
			Response Response `json:"data"`
		}
	)
	tempPath = "./static/" + r.Name

	// TODO
	err = f(r.Path, tempPath)
	if nil != err {
		return nil, err
	}

	attachments, _ := os.Open(tempPath)
	defer func() {
		err := attachments.Close()
		if nil != err {
			log.Errorf("failed to close temporary file: %s", err.Error())
		}

		err = os.Remove(tempPath)
		if nil != err {
			log.Errorf("failed to remove temporary file: %s", err.Error())
		}
	}()

	// Step 1. Try to determine the contentType.
	contentType, body, err := fDetectContentType(attachments)
	if nil != err {
		log.Errorf("failed to detect content type: %s", err.Error())
		return
	}

	if nil == body {
		log.Errorf("empty attachment")
		return nil, errors.New("generating attachment failed")
	}

	// Step 2:
	// Initiate and then make the upload.
	prc, pwc := io.Pipe()
	mpartW := multipart.NewWriter(pwc)
	go func() {
		defer func() {
			_ = mpartW.Close()
			_ = pwc.Close()
		}()
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes("file"), escapeQuotes(r.Name)))
		//h.Set("Content-Type", "application/octet-stream")
		log.Infof("Content-Type: %s", contentType)
		h.Set("Content-Type", contentType)
		formFile, err := mpartW.CreatePart(h)
		//formFile, err := mpartW.CreateFormFile("file", r.Name)
		if err != nil {
			return
		}
		_, _ = io.Copy(formFile, body)
		//writeStringField(mpartW, "Content-Type", contentType)
		//writeStringField(mpartW, "resource_subtype", "asana")
		//writeStringField(mpartW, "name", "20210811065901")
	}()

	fullURL := fmt.Sprintf("%s/api/1.0/tasks/%s/attachments", util.AsanaHost, r.TaskID)

	req, err := http.NewRequest("POST", fullURL, prc)
	if err != nil {
		log.Errorf("failed to build attachment request: %s", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", mpartW.FormDataContentType())
	req.Header.Set("Authorization", r.PAToken)

	buf, err := util.Request(req)
	if nil != err {
		return nil, err
	}
	err = json.Unmarshal(buf, &data)
	if nil != err {
		log.Errorf("[%s][%s]failed to upload attachments: %s", r.TicketID, r.TaskID, err.Error())
		return nil, err
	}
	return &data.Response, err
}

func fDetectContentType(r io.Reader) (string, io.Reader, error) {
	if r == nil {
		return "", nil, errors.New("empty attachments")
	}

	seeker, seekable := r.(io.Seeker)
	sniffBuf := make([]byte, 512)
	n, err := io.ReadAtLeast(r, sniffBuf, 1)
	if err != nil {
		log.Errorf(err.Error())
		return "", nil, err
	}

	contentType := http.DetectContentType(sniffBuf)
	needsRepad := !seekable
	if seekable {
		if _, err = seeker.Seek(int64(-n), io.SeekCurrent); err != nil {
			// Since we failed to rewind it, mark it as needing repad
			needsRepad = true
		}
	}

	if needsRepad {
		r = io.MultiReader(bytes.NewReader(sniffBuf), r)
	}

	return contentType, r, nil
}

func writeStringField(w *multipart.Writer, key, value string) {
	fw, err := w.CreateFormField(key)
	if err == nil {
		_, _ = io.WriteString(fw, value)
	}
}

var (
	errValidation = errors.New("invalid request")
)

func download(src, filename string) error {
	if "" == src || "" == filename {
		return errValidation
	}
	f, err := os.Create(filename)
	if nil != err {
		log.Errorf("failed to create attachments")
		return errors.New("failed to create file")
	}
	defer f.Close()
	//defer os.Remove(filename)

	resp, err := http.Get(src)
	if nil != err {
		log.Infof("download report failed: %s", err.Error())
		return err
	}

	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
