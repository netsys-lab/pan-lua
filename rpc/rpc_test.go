// Copyright 2021 Thorben Krüger (thorben.krueger@ovgu.de)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rpc

import (
	"testing"
)

func TestIDServer(t *testing.T) {
	s := IDServer{42}
	var id IDMsg
	s.GetID(new(IDMsg), &id)
	if *id.Value != 43 {
		t.Errorf("GetID Value = %d, want 43", *id.Value)
	}
	s.GetID(new(IDMsg), &id)
	if *id.Value != 44 {
		t.Errorf("GetID Value = %d, want 44", *id.Value)
	}

}
