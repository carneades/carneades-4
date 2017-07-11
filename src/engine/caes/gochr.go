// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Wrapper for the GoCHR inference engine

package caes

import (
	"fmt"
	// "log"
	"strings"

	chr "github.com/hfried/GoCHR/src/engine/CHR"
)

// Translate a theory into a GoCHR rulestore
func TheoryToRuleStore(t *Theory) *chr.RuleStore {
	// log.Printf("TheoryToRuleStore\n") // DEBUG
	rs := chr.MakeRuleStore()
	for _, s := range t.ArgSchemes {
		// If the scheme has no conclusions, skip the scheme
		// and assume it only defines a weighing function but no rule
		if len(s.Conclusions) > 0 {
			// A "go" term is added to CHR rules for
			// argument schemes with no premises, since CHR requires
			// rules to have at least one term in the head.
			var premises []string
			if len(s.Premises) > 0 {
				premises = s.Premises
			} else if len(s.Deletions) == 0 {
				premises = []string{"go"}
			}
			argTerm := fmt.Sprintf("argument(%s,[%s])", s.Id, strings.Join(s.Variables, ","))
			conclusions := append(s.Conclusions, s.Assumptions...)
			conclusions = append(conclusions, argTerm)
			// Errors raised by AddRule are ignored. The rule is just skipped.
			// Note that the body of the rules includes the assumptions
			// and conclusions of the scheme.
			// fmt.Printf("AddRule id=%s, premises=%s, deletions=%s, guards=%s, conclusions=%s\n", s.Id, premises, s.Deletions, s.Guards, conclusions)
			rs.AddRule(s.Id, premises, s.Deletions, s.Guards, conclusions)
		}
	}
	return rs
}
