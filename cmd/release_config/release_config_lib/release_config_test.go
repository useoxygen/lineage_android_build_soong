// Copyright 2026 Google Inc. All rights reserved.
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

package release_config_lib

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	rc_proto "android/soong/cmd/release_config/release_config_proto"
)

func TestWriteMakefileRetainsUnchangedMtime(t *testing.T) {
	config := &ReleaseConfig{
		Name:                  "canary",
		ReleaseConfigType:     rc_proto.ReleaseConfigType_RELEASE_CONFIG,
		ReleaseConfigArtifact: &rc_proto.ReleaseConfigArtifact{},
		FlagArtifacts:         FlagArtifacts{},
	}
	configs := &ReleaseConfigs{ReleaseConfigs: map[string]*ReleaseConfig{"canary": config}}
	path := filepath.Join(t.TempDir(), "release.vars")
	write := func() {
		t.Helper()
		if err := config.WriteMakefile(path, "canary", configs); err != nil {
			t.Fatal(err)
		}
	}
	read := func() []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	mtime := func() time.Time {
		t.Helper()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		return info.ModTime()
	}
	fixed := time.Unix(1234567890, 0)
	preserve := func() {
		t.Helper()
		before := read()
		if err := os.Chtimes(path, fixed, fixed); err != nil {
			t.Fatal(err)
		}
		expected := mtime()
		write()
		if !bytes.Equal(read(), before) {
			t.Fatal("identical rendering changed file contents")
		}
		if !mtime().Equal(expected) {
			t.Fatal("identical rendering changed file modification time")
		}
	}
	write()
	original := read()
	if !bytes.Contains(original, []byte("ALL_RELEASE_CONFIGS_FOR_PRODUCT :=$= canary")) {
		t.Fatal("new makefile is missing its release configuration")
	}
	preserve()
	config.DisallowLunchUse = true
	write()
	changed := read()
	if bytes.Equal(changed, original) || !bytes.Contains(changed, []byte("_disallow_lunch_use :=$= true")) {
		t.Fatal("changed configuration did not replace the rendered contents")
	}
	if mtime().Equal(fixed) {
		t.Fatal("changed rendering retained the old modification time")
	}
	preserve()
}

func TestWriteMakefileReportsWriteError(t *testing.T) {
	config := &ReleaseConfig{Name: "canary", ReleaseConfigArtifact: &rc_proto.ReleaseConfigArtifact{}, FlagArtifacts: FlagArtifacts{}}
	configs := &ReleaseConfigs{ReleaseConfigs: map[string]*ReleaseConfig{}}
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("blocked"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := config.WriteMakefile(filepath.Join(parent, "release.vars"), "canary", configs); err == nil {
		t.Fatal("unwritable output path succeeded")
	}
}
