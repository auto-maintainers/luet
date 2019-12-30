// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>
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

package pkg_test

import (
	. "github.com/mudler/luet/pkg/package"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Versions", func() {

	Context("Versions Parser1", func() {
		v, err := ParseVersion(">=1.0")
		It("ParseVersion1", func() {
			var c PkgSelectorCondition = PkgCondGreaterEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser2", func() {
		v, err := ParseVersion(">1.0")
		It("ParseVersion2", func() {
			var c PkgSelectorCondition = PkgCondGreater
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser3", func() {
		v, err := ParseVersion("<=1.0")
		It("ParseVersion3", func() {
			var c PkgSelectorCondition = PkgCondLessEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser4", func() {
		v, err := ParseVersion("<1.0")
		It("ParseVersion4", func() {
			var c PkgSelectorCondition = PkgCondLess
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser5", func() {
		v, err := ParseVersion("=1.0")
		It("ParseVersion5", func() {
			var c PkgSelectorCondition = PkgCondEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser6", func() {
		v, err := ParseVersion("!1.0")
		It("ParseVersion6", func() {
			var c PkgSelectorCondition = PkgCondNot
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("1.0"))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser7", func() {
		v, err := ParseVersion("")
		It("ParseVersion7", func() {
			var c PkgSelectorCondition = PkgCondInvalid
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal(""))
			Expect(v.VersionSuffix).Should(Equal(""))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser8", func() {
		v, err := ParseVersion("=12.1.0.2_p1")
		It("ParseVersion8", func() {
			var c PkgSelectorCondition = PkgCondEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("12.1.0.2"))
			Expect(v.VersionSuffix).Should(Equal("_p1"))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser9", func() {
		v, err := ParseVersion(">=0.0.20190406.4.9.172-r1")
		It("ParseVersion9", func() {
			var c PkgSelectorCondition = PkgCondGreaterEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("0.0.20190406.4.9.172"))
			Expect(v.VersionSuffix).Should(Equal("-r1"))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Versions Parser10", func() {
		v, err := ParseVersion(">=0.0.20190406.4.9.172_alpha")
		It("ParseVersion10", func() {
			var c PkgSelectorCondition = PkgCondGreaterEqual
			Expect(err).Should(BeNil())
			Expect(v.Version).Should(Equal("0.0.20190406.4.9.172"))
			Expect(v.VersionSuffix).Should(Equal("_alpha"))
			Expect(v.Condition).Should(Equal(c))
		})
	})

	Context("Selector1", func() {
		v1, err := ParseVersion(">=0.0.20190406.4.9.172-r1")
		v2, err2 := ParseVersion("1.0.111")
		match, err3 := PackageAdmit(v1, v2)
		It("Selector1", func() {
			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())
			Expect(err3).Should(BeNil())
			Expect(match).Should(Equal(true))
		})
	})

	Context("Selector2", func() {
		v1, err := ParseVersion(">=0.0.20190406.4.9.172-r1")
		v2, err2 := ParseVersion("0")
		match, err3 := PackageAdmit(v1, v2)
		It("Selector2", func() {
			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())
			Expect(err3).Should(BeNil())
			Expect(match).Should(Equal(false))
		})
	})

	Context("Selector3", func() {
		v1, err := ParseVersion(">0")
		v2, err2 := ParseVersion("0.0.40-alpha")
		match, err3 := PackageAdmit(v1, v2)
		It("Selector3", func() {
			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())
			Expect(err3).Should(BeNil())
			Expect(match).Should(Equal(true))
		})
	})

	Context("Selector4", func() {
		v1, err := ParseVersion(">0")
		v2, err2 := ParseVersion("")
		match, err3 := PackageAdmit(v1, v2)
		It("Selector4", func() {
			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())
			Expect(err3).Should(BeNil())
			Expect(match).Should(Equal(true))
		})
	})
})
