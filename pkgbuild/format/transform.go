package format

import (
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
)

type TransformFunc func(*info.NodeInfo) *info.NodeInfo

type TransformListFunc func(info.NodeInfoList) info.NodeInfoList

func MultiTransform(tranforms ...TransformFunc) TransformFunc {
	return func(node *info.NodeInfo) *info.NodeInfo {
		for _, transform := range tranforms {
			node = transform(node)
		}
		return node
	}
}

func TransformAndRecomputePositions(
	node *info.NodeInfo,
	transform TransformFunc,
	diff *position.PosDiff,
) (*info.NodeInfo, *position.PosDiff) {
	_, nextFromPos := node.Position()
	transformedNode := transform(node)

	diff.Update(transformedNode.Stmt)

	_, newEnd := transformedNode.Position()
	nextDiff := position.Diff(nextFromPos, newEnd)

	return transformedNode, nextDiff
}

func TransformList(nodes info.NodeInfoList, tranforms ...TransformFunc) info.NodeInfoList {
	var diff *position.PosDiff
	transform := MultiTransform(tranforms...)

	for i, node := range nodes {
		nodes[i], diff = TransformAndRecomputePositions(node, transform, diff)
	}

	return nodes
}
