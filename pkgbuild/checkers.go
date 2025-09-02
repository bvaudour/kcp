package pkgbuild

import (
	"slices"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/standard"
	"git.kaosx.ovh/benjamin/collection"
)

// HasHeader returns true if the PKGBUILD has leading comments.
func (p *PKGBUILD) HasHeader() bool {
	if len(p.NodeInfoList) > 0 {
		n := p.NodeInfoList[0]
		begin, _ := n.InnerPosition()
		return begin.Line() > 1 || len(n.HeaderComments()) > 0
	}

	return false
}

// GetMissingVariables returns the missing standard variables names of the PKGBUILD
// and if it has checksums.
func (p *PKGBUILD) GetMissingVariables() (missing []string, missingChecksum bool) {
	variables := collection.NewSet(p.GetVariables()...)
	for _, required := range standard.GetRequiredVariables() {
		if !variables.Contains(required) {
			missing = append(missing, required)
		}
	}

	missingChecksum = !slices.ContainsFunc(
		standard.GetChecksumsVariables(),
		func(chck string) bool { return variables.Contains(chck) },
	)

	return
}

// GetMissingFunctions returns the missing standard functions of the PKGBUILD.
func (p *PKGBUILD) GetMissingFunctions() (missing []string) {
	functions := collection.NewSet(p.GetFunctions()...)
	for _, required := range standard.GetRequiredFunctions() {
		if !functions.Contains(required) {
			missing = append(missing, required)
		}
	}

	return
}

// GetBadStandard returns the nodes which are not on the good type (function declaration, single variable or array variable).
func (p *PKGBUILD) GetBadStandard() (badInfos info.NodeInfoList, shouldBe []info.NodeType) {
	for _, n := range p.NodeInfoList {
		var t info.NodeType
		if standard.IsStandardFunction(n.Name) {
			t = info.Function
		} else if standard.IsStandardVariable(n.Name) {
			if standard.IsArrayVariable(n.Name) {
				t = info.ArrayVar
			} else {
				t = info.SingleVar
			}
		}
		if t != info.Unknown && n.Type != t {
			badInfos, shouldBe = append(badInfos, n), append(shouldBe, t)
		}
	}

	return
}

// GetEmpty returns the variables which have no value.
func (p *PKGBUILD) GetEmpty() (infos info.NodeInfoList) {
	for _, n := range p.NodeInfoList {
		if (n.Type == info.ArrayVar && len(n.Values) == 0) || (n.Type == info.SingleVar && len(n.Value) == 0) {
			infos = append(infos, n)
		}
	}

	return
}

// IsPkgrelClean returns true if pkgrel = 1.
func (p *PKGBUILD) IsPkgrelClean() bool {
	return p.GetValue(standard.PKGREL) == "1"
}

// IsArchClean returns true if arch contains only x86_64.
func (p *PKGBUILD) IsArchClean() bool {
	arch := p.GetArrayValue(standard.ARCH)

	return len(arch) == 1 && arch[0] == "x86_64"
}

// HadDepends returns true if depends, makedepends or checkdepends is defined.
func (p *PKGBUILD) HasDepends() bool {
	for _, k := range []string{standard.DEPENDS, standard.MAKEDEPENDS, standard.CHECKDEPENDS} {
		v := p.GetArrayValue(k)
		if len(v) > 0 {
			return true
		}
	}

	return false
}

// IsInstallValid return true if install refers to existent files.
func (p *PKGBUILD) IsInstallValid() bool {
	return !p.HasValue(standard.INSTALL) || common.FileExists(p.GetValue(standard.INSTALL))
}
