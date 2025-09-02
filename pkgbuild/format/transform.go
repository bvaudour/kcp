package format

import (
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
)

// TransformFunc represents a function to modify a node.
type TransformFunc func(*info.NodeInfo) *info.NodeInfo

// TransformListFunc represents a function to modify a list of nodes.
type TransformListFunc func(info.NodeInfoList) info.NodeInfoList

// MultiTransform fusions many node's transform functions in one.
func MultiTransform(tranforms ...TransformFunc) TransformFunc {
	return func(node *info.NodeInfo) *info.NodeInfo {
		for _, transform := range tranforms {
			node = transform(node)
		}
		return node
	}
}

// TransformAndRecomputePositions applies a transform function of a node,
// and moves its position based on diff.
// Its returns the transformed node and the new diff position after transformation.
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

// TransformList transforms all nodes of a list and recomputes positions.
func TransformList(nodes info.NodeInfoList, tranforms ...TransformFunc) info.NodeInfoList {
	var diff *position.PosDiff
	transform := MultiTransform(tranforms...)

	for i, node := range nodes {
		nodes[i], diff = TransformAndRecomputePositions(node, transform, diff)
	}

	return nodes
}
