//  Copyright 2019 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package policies

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	agentendpointpb "google.golang.org/genproto/googleapis/cloud/osconfig/agentendpoint/v1beta"
)

func runYumRepositories(ctx context.Context, repos []*agentendpointpb.YumRepository) (string, error) {
	td, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %v", err)
	}
	defer os.RemoveAll(td)
	testRepo := filepath.Join(td, "testRepo")

	if err := yumRepositories(ctx, repos, testRepo); err != nil {
		return "", fmt.Errorf("error running yumRepositories: %v", err)
	}

	data, err := ioutil.ReadFile(testRepo)
	if err != nil {
		return "", fmt.Errorf("error reading testRepo: %v", err)
	}

	return string(data), nil
}

func TestYumRepositories(t *testing.T) {
	tests := []struct {
		desc  string
		repos []*agentendpointpb.YumRepository
		want  string
	}{
		{"no repos", []*agentendpointpb.YumRepository{}, "# Repo file managed by Google OSConfig agent\n[main]\ngpgcheck=1\n"},
		{
			"1 repo",
			[]*agentendpointpb.YumRepository{
				{BaseUrl: "http://repo1-url/", Id: "id"},
			},
			"# Repo file managed by Google OSConfig agent\n[main]\ngpgcheck=1\n\n[id]\nname=id\nbaseurl=http://repo1-url/\nenabled=1\ngpgcheck=1\n",
		},
		{
			"2 repos",
			[]*agentendpointpb.YumRepository{
				{BaseUrl: "http://repo1-url/", Id: "id1", DisplayName: "displayName1", GpgKeys: []string{"https://url/key"}},
				{BaseUrl: "http://repo1-url/", Id: "id2", DisplayName: "displayName2", GpgKeys: []string{"https://url/key1", "https://url/key2"}},
			},
			"# Repo file managed by Google OSConfig agent\n[main]\ngpgcheck=1\n\n[id1]\nname=displayName1\nbaseurl=http://repo1-url/\nenabled=1\ngpgcheck=1\ngpgkey=https://url/key\n\n[id2]\nname=displayName2\nbaseurl=http://repo1-url/\nenabled=1\ngpgcheck=1\ngpgkey=https://url/key1\n       https://url/key2\n",
		},
	}

	for _, tt := range tests {
		got, err := runYumRepositories(context.Background(), tt.repos)
		if err != nil {
			t.Fatal(err)
		}

		if got != tt.want {
			t.Errorf("%s: got:\n%q\nwant:\n%q", tt.desc, got, tt.want)
		}
	}
}
