// Copyright 2015 Matthew Holt and The Caddy Authors
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

package caddyzstd

import (
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/zstd"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/encode"
)

func init() {
	caddy.RegisterModule(Zstd{})
}

// Zstd can create Zstandard encoders.
type Zstd struct {
	// Compression level refer to type constants value from zstd.SpeedFastest to zstd.SpeedBestCompression
	Level zstd.EncoderLevel `json:"level,omitempty"`

	WindowSize *uint64 `json:"window_size"`
}

// CaddyModule returns the Caddy module information.
func (Zstd) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.encoders.zstd",
		New: func() caddy.Module { return new(Zstd) },
	}
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens.
func (z *Zstd) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume option name
	if !d.NextArg() {
		return nil
	}
	levelStr := d.Val()
	ok, level := zstd.EncoderLevelFromString(levelStr)

	if !ok {
		return d.Errf("unexpected compression level, use one of '%s', '%s', '%s', '%s'",
			zstd.SpeedFastest,
			zstd.SpeedDefault,
			zstd.SpeedBetterCompression,
			zstd.SpeedBestCompression,
		)
	}
	z.Level = level

	if !d.NextArg() {
		return nil
	}
	windowStr := d.Val()
	size, err := humanize.ParseBytes(windowStr)
	if err != nil {
		return d.Errf("incorrect window size: %v", err)
	}
	z.WindowSize = &size

	return nil
}

// Provision provisions g's configuration.
func (z *Zstd) Provision(ctx caddy.Context) error {
	if z.WindowSize == nil {
		// The default of 8MB for the window is
		// too large for many clients, so we limit
		// it to 128K to lighten their load.
		ws := uint64(128 << 10)
		z.WindowSize = &ws
	}
	return nil
}

// AcceptEncoding returns the name of the encoding as
// used in the Accept-Encoding request headers.
func (Zstd) AcceptEncoding() string { return "zstd" }

// NewEncoder returns a new Zstandard writer.
func (z Zstd) NewEncoder() encode.Encoder {
	opts := []zstd.EOption{
		zstd.WithEncoderConcurrency(1),
		zstd.WithZeroFrames(true),
		zstd.WithEncoderLevel(z.Level),
	}
	if *z.WindowSize > 0 {
		opts = append(opts, zstd.WithWindowSize(int(*z.WindowSize)))
	}
	writer, _ := zstd.NewWriter(nil, opts...)
	return writer
}

// Interface guards
var (
	_ encode.Encoding       = (*Zstd)(nil)
	_ caddyfile.Unmarshaler = (*Zstd)(nil)
	_ caddy.Provisioner     = (*Zstd)(nil)
)
