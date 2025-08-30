package format

import (
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"mvdan.cc/sh/v3/syntax"
)

type FormatOption uint

const (
	OptionRemoveOuterComments FormatOption = 1 << iota
	OptionRemoveInnerComments
	OptionRemoveHeader
	OptionRemoveDuplicates
	OptionFormatWords
	OptionIndentFunctionsWithSpaces
	OptionIndentFunctionsWithTabs
	OptionIndentArrayVariables
	OptionReorder
	OptionFormatBlank
	OptionKeepFirstBlank

	OptionIndentFunctions  = OptionIndentFunctionsWithSpaces | OptionIndentFunctionsWithTabs
	OptionIndentWithSpaces = OptionIndentArrayVariables | OptionIndentFunctionsWithSpaces
	OptionIndentWithTabs   = OptionIndentArrayVariables | OptionIndentFunctionsWithTabs
)

func (option FormatOption) Merge(options ...FormatOption) FormatOption {
	for _, o := range options {
		option |= o
	}

	return option
}

func (option FormatOption) Contains(arg FormatOption) bool {
	return option&arg != 0
}

type Formater interface {
	Format(info.NodeInfoList, []syntax.Comment) (info.NodeInfoList, []syntax.Comment)
}

type formater struct {
	RemoveHeader        bool
	RemoveDuplicates    bool
	RemoveOuterComments bool
	RemoveInnerComments bool
	FormatWords         bool
	IndentFunctions     bool
	IndentVariables     bool
	IndentSpaces        uint
	Reorder             bool
	FormatBlank         bool
	KeepFirstBlank      bool
}

func (f formater) Format(nodes info.NodeInfoList, lastComments []syntax.Comment) (newNodes info.NodeInfoList, newComments []syntax.Comment) {
	l := len(nodes)
	hasNodes := l > 0
	if !hasNodes {
		if !f.RemoveHeader && f.RemoveOuterComments {
			newComments = lastComments
		}
		return
	}

	oldBegin, oldEnd := nodes[l-1].Position()

	if f.RemoveHeader {
		nodes = RemoveHeader(nodes)
	}

	if f.RemoveDuplicates {
		nodes = RemoveDuplicates(nodes)
	}

	var transforms []TransformFunc
	if f.RemoveOuterComments {
		transforms = append(transforms, RemoveOuterComments)
	}
	if f.RemoveInnerComments {
		transforms = append(transforms, RemoveInnerComments)
	}
	if f.FormatWords {
		transforms = append(transforms, FormatWords)
	}
	if f.IndentFunctions {
		transforms = append(transforms, IndentFunctions(f.IndentSpaces))
	}
	if f.IndentVariables {
		transforms = append(transforms, IndentVariables)
	}

	newNodes = TransformList(nodes, transforms...)
	if f.Reorder {
		newNodes = Reorder(nodes)
	}
	if f.FormatBlank {
		newNodes = FormatBlankLines(f.KeepFirstBlank)(nodes)
	}

	if !f.RemoveOuterComments && len(lastComments) > 0 {
		var diff *position.PosDiff
		if l := len(newNodes); l == 0 {
			newBegin := lastComments[0].Pos()
			diff = position.Diff(oldBegin, newBegin)
		} else {
			_, newEnd := newNodes[l-1].Position()
			diff = position.Diff(oldEnd, newEnd)
		}
		newComments = make([]syntax.Comment, len(lastComments))
		copy(newComments, lastComments)
		diff.UpdateComments(newComments)
	}

	return
}

func NewFormater(options ...FormatOption) Formater {
	var option FormatOption
	option = option.Merge(options...)

	f := formater{
		RemoveHeader:        option.Contains(OptionRemoveHeader),
		RemoveDuplicates:    option.Contains(OptionRemoveDuplicates),
		RemoveOuterComments: option.Contains(OptionRemoveOuterComments),
		RemoveInnerComments: option.Contains(OptionRemoveInnerComments),
		FormatWords:         option.Contains(OptionFormatWords),
		IndentFunctions:     option.Contains(OptionIndentFunctions),
		IndentVariables:     option.Contains(OptionIndentArrayVariables),
		Reorder:             option.Contains(OptionReorder),
		FormatBlank:         option.Contains(OptionFormatBlank),
		KeepFirstBlank:      option.Contains(OptionKeepFirstBlank),
	}

	if option.Contains(OptionIndentFunctionsWithSpaces) {
		f.IndentSpaces = 2
	}

	return f
}
