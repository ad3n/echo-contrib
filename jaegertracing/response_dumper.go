// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2017 LabStack and Echo contributors

package jaegertracing

import (
	"bytes"
	"io"
	"net/http"
)

type responseDumper struct {
	http.ResponseWriter

	buf bytes.Buffer
}

func newResponseDumper(resp http.ResponseWriter) *responseDumper {
	return &responseDumper{
		ResponseWriter: resp,
	}
}

func (d *responseDumper) Write(b []byte) (int, error) {
	n, err := d.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}

	if n != len(b) {
		return n, io.ErrShortWrite
	}

	return d.buf.Write(b)
}

func (d *responseDumper) GetResponse() string {
	return d.buf.String()
}

func (d *responseDumper) Unwrap() http.ResponseWriter {
	return d.ResponseWriter
}
