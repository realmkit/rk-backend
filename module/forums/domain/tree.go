package domain

import structuremodel "github.com/niflaot/gamehub-go/module/forums/domain/structure"

// ForumStats stores denormalized forum counters and latest post summary.
type ForumStats = structuremodel.ForumStats

// ForumNode is one visible forum tree node.
type ForumNode = structuremodel.ForumNode

// CategoryNode is one visible category with forums.
type CategoryNode = structuremodel.CategoryNode

// ForumTree is the forum home tree response.
type ForumTree = structuremodel.ForumTree
