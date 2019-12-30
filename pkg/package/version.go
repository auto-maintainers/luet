// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>,
//                  Daniele Rondina <geaaru@sabayonlinux.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

package pkg

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"
)

// Package Selector Condition
type PkgSelectorCondition int

type PkgVersionSelector struct {
	Version       string
	VersionSuffix string
	Condition     PkgSelectorCondition
	// TODO: Integrate support for multiple repository
}

const (
	PkgCondInvalid = 0
	// >
	PkgCondGreater = 1
	// >=
	PkgCondGreaterEqual = 2
	// <
	PkgCondLess = 3
	// <=
	PkgCondLessEqual = 4
	// =
	PkgCondEqual = 5
	// !
	PkgCondNot = 6
	// ~
	PkgCondAnyRevision = 7
	// =<pkg>*
	PkgCondMatchVersion = 8
)

func ParseVersion(v string) (PkgVersionSelector, error) {
	var ans PkgVersionSelector = PkgVersionSelector{
		Version:       "",
		VersionSuffix: "",
		Condition:     PkgCondInvalid,
	}

	if strings.HasPrefix(v, ">=") {
		v = v[2:]
		ans.Condition = PkgCondGreaterEqual
	} else if strings.HasPrefix(v, ">") {
		v = v[1:]
		ans.Condition = PkgCondGreater
	} else if strings.HasPrefix(v, "<=") {
		v = v[2:]
		ans.Condition = PkgCondLessEqual
	} else if strings.HasPrefix(v, "<") {
		v = v[1:]
		ans.Condition = PkgCondLess
	} else if strings.HasPrefix(v, "=") {
		v = v[1:]
		if strings.HasSuffix(v, "*") {
			ans.Condition = PkgCondMatchVersion
			v = v[0 : len(v)-1]
		} else {
			ans.Condition = PkgCondEqual
		}
	} else if strings.HasPrefix(v, "~") {
		v = v[1:]
		ans.Condition = PkgCondAnyRevision
	} else if strings.HasPrefix(v, "!") {
		v = v[1:]
		ans.Condition = PkgCondNot
	}

	regexPkg := regexp.MustCompile(
		fmt.Sprintf("(%s|%s|%s|%s|%s|%s)((%s|%s|%s|%s|%s|%s|%s)+)*$",
			// Version regex
			// 1.1
			"[0-9]+[.][0-9]+[a-z]*",
			// 1
			"[0-9]+[a-z]*",
			// 1.1.1
			"[0-9]+[.][0-9]+[.][0-9]+[a-z]*",
			// 1.1.1.1
			"[0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[a-z]*",
			// 1.1.1.1.1
			"[0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[a-z]*",
			// 1.1.1.1.1.1
			"[0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[.][0-9]+[a-z]*",
			// suffix
			"-r[0-9]+",
			"_p[0-9]+",
			"_pre[0-9]*",
			"_rc[0-9]+",
			// handle also rc without number
			"_rc",
			"_alpha",
			"_beta",
		),
	)
	matches := regexPkg.FindAllString(v, -1)

	if len(matches) > 0 {
		// Check if there patch
		if strings.Contains(matches[0], "_p") {
			ans.Version = matches[0][0:strings.Index(matches[0], "_p")]
			ans.VersionSuffix = matches[0][strings.Index(matches[0], "_p"):]
		} else if strings.Contains(matches[0], "_rc") {
			ans.Version = matches[0][0:strings.Index(matches[0], "_rc")]
			ans.VersionSuffix = matches[0][strings.Index(matches[0], "_rc"):]
		} else if strings.Contains(matches[0], "_alpha") {
			ans.Version = matches[0][0:strings.Index(matches[0], "_alpha")]
			ans.VersionSuffix = matches[0][strings.Index(matches[0], "_alpha"):]
		} else if strings.Contains(matches[0], "_beta") {
			ans.Version = matches[0][0:strings.Index(matches[0], "_beta")]
			ans.VersionSuffix = matches[0][strings.Index(matches[0], "_beta"):]
		} else if strings.Contains(matches[0], "-r") {
			ans.Version = matches[0][0:strings.Index(matches[0], "-r")]
			ans.VersionSuffix = matches[0][strings.Index(matches[0], "-r"):]
		} else {
			ans.Version = matches[0]
		}
	}

	// Set condition if there isn't a prefix but only a version
	if ans.Condition == PkgCondInvalid && ans.Version != "" {
		ans.Condition = PkgCondEqual
	}

	// NOTE: Now suffix complex like _alpha_rc1 are not supported.
	return ans, nil
}

func PackageAdmit(selector, i PkgVersionSelector) (bool, error) {
	var v1 *version.Version = nil
	var v2 *version.Version = nil
	var ans bool
	var err error

	if selector.Version != "" {
		v1, err = version.NewVersion(selector.Version)
		if err != nil {
			return false, err
		}
	}
	if i.Version != "" {
		v2, err = version.NewVersion(i.Version)
		if err != nil {
			return false, err
		}
	} else {
		// If version is not defined match always package
		ans = true
	}

	// If package doesn't define version admit all versions of the package.
	if selector.Version == "" {
		ans = true
	} else {
		if selector.Condition == PkgCondInvalid || selector.Condition == PkgCondEqual {
			// case 1: source-pkg-1.0 and dest-pkg-1.0 or dest-pkg without version
			if i.Version != "" && i.Version == selector.Version && selector.VersionSuffix == i.VersionSuffix {
				ans = true
			}
		} else if selector.Condition == PkgCondAnyRevision {
			if v1 != nil && v2 != nil {
				ans = v1.Equal(v2)
			}
		} else if selector.Condition == PkgCondMatchVersion {
			// TODO: case of 7.3* where 7.30 is accepted.
			if v1 != nil && v2 != nil {
				segments := v1.Segments()
				n := strings.Count(selector.Version, ".")
				switch n {
				case 0:
					segments[0]++
				case 1:
					segments[1]++
				case 2:
					segments[2]++
				default:
					segments[len(segments)-1]++
				}
				nextVersion := strings.Trim(strings.Replace(fmt.Sprint(segments), " ", ".", -1), "[]")
				constraints, err := version.NewConstraint(
					fmt.Sprintf(">= %s, < %s", selector.Version, nextVersion),
				)
				if err != nil {
					return false, err
				}
				ans = constraints.Check(v2)
			}
		} else if v1 != nil && v2 != nil {

			// TODO: Integrate check of version suffix
			switch selector.Condition {
			case PkgCondGreaterEqual:
				ans = v2.GreaterThanOrEqual(v1)
			case PkgCondLessEqual:
				ans = v2.LessThanOrEqual(v1)
			case PkgCondGreater:
				ans = v2.GreaterThan(v1)
			case PkgCondLess:
				ans = v2.LessThan(v1)
			case PkgCondNot:
				ans = !v2.Equal(v1)
			}
		}
	}

	return ans, nil
}

func (p *DefaultPackage) SelectorMatchVersion(v string) (bool, error) {
	if !p.IsSelector() {
		return false, errors.New("Package is not a selector")
	}

	vS, err := ParseVersion(p.GetVersion())
	if err != nil {
		return false, err
	}

	vSI, err := ParseVersion(v)
	if err != nil {
		return false, err
	}

	return PackageAdmit(vS, vSI)
}

func (p *DefaultPackage) VersionMatchSelector(selector string) (bool, error) {
	vS, err := ParseVersion(selector)
	if err != nil {
		return false, err
	}

	vSI, err := ParseVersion(p.GetVersion())
	if err != nil {
		return false, err
	}

	return PackageAdmit(vS, vSI)
}
