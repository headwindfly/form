// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package form is a form decoder that decode request body into a struct.
package form

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/gorilla/schema"
)

// Content type constants.
const (
	ContentType              = "Content-Type"
	ContentTypeForm          = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm = "multipart/form-data"
	ContentTypeJSON          = "application/json"
	ContentTypeXML           = "application/xml"
)

var decoders = make(map[string]Decoder)

var defaultDecoder = schema.NewDecoder()
var defaultMaxMemory int64 = 10 * 1024 * 1024

func init() {
}

func init() {
	defaultDecoder.IgnoreUnknownKeys(true)
	defaultDecoder.SetAliasTag("json")

	Register(ContentTypeForm, NewForm(defaultDecoder))
	Register(ContentTypeMultipartForm, NewMultipartForm(defaultMaxMemory))
	Register(ContentTypeJSON, JSON)
	Register(ContentTypeXML, XML)
}

// Register a decoder for the given content type.
func Register(contentType string, decoder Decoder) {
	decoders[contentType] = decoder
}

// Decoder is a function that decode data from request into v.
type Decoder func(req *http.Request, v interface{}) error

// Decode data from a request into v, v should be a pointer.
func Decode(r *http.Request, v interface{}) error {
	contentType, err := parseContentType(r)
	if err != nil {
		return err
	}
	if decoder, ok := decoders[contentType]; ok {
		return decoder(r, v)
	}

	return errors.New("Unsupported content type: " + contentType)
}

func parseContentType(r *http.Request) (string, error) {
	header := r.Header.Get(ContentType)
	contentType, _, err := mime.ParseMediaType(header)
	if err != nil {
		return "", err
	}

	return contentType, nil
}

// JSON is a JSON decoder.
func JSON(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// XML is an XML decoder.
func XML(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return xml.Unmarshal(body, v)
}

// NewForm returns a post form decoder with the given schema decoder.
func NewForm(decoder *schema.Decoder) Decoder {
	return func(r *http.Request, v interface{}) error {
		err := r.ParseForm()
		if err != nil {
			return err
		}

		return decoder.Decode(v, r.PostForm)
	}
}

// NewMultipartForm returns a multipart form decoder with the given schema decoder.
func NewMultipartForm(maxMemory int64) Decoder {
	return func(r *http.Request, v interface{}) error {
		err := r.ParseMultipartForm(maxMemory)
		if err != nil {
			return err
		}

		return defaultDecoder.Decode(v, r.MultipartForm.Value)
	}
}
