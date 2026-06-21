package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
)

// GetInheritanceTree builds a role inheritance tree with permissions
func (uc *PermissionUseCase) GetInheritanceTree(ctx context.Context) (*model.InheritanceTreeResponse, error) {
	// Get all roles
	roles, err := uc.RoleRepo.FindAll(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get roles: %v", err)
		return nil, err
	}

	// Get all permissions
	allPerms, err := uc.GetAllPermissions(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get all permissions: %v", err)
		return nil, err
	}

	// Build role nodes map
	roleNodesMap := make(map[string]*model.RoleNode)
	parentMap := make(map[string][]string) // child -> parents

	// First pass: Create all role nodes and identify parent relationships
	for _, role := range roles {
		node := &model.RoleNode{
			ID:                   role.ID,
			Name:                 role.Name,
			Description:          role.Description,
			Children:             []model.RoleNode{},
			OwnPermissions:       [][]string{},
			InheritedPermissions: [][]string{},
			EffectivePermissions: [][]string{},
			Parents:              []string{},
		}

		// Get parent roles (role inheritance via Casbin grouping)
		parents, err := uc.GetParentRoles(ctx, role.Name, "global")
		if err == nil && len(parents) > 0 {
			parentMap[role.Name] = parents
			node.Parents = parents
			// For simple tree UI compatibility, set first parent as ParentID
			parentName := parents[0]
			node.ParentID = &parentName
		}

		roleNodesMap[role.Name] = node
	}

	// Second pass: Assign permissions to roles
	for _, perm := range allPerms {
		if len(perm) < 4 {
			continue
		}

		roleName := perm[0]
		// perm format: [role, domain, path, method]

		if node, exists := roleNodesMap[roleName]; exists {
			node.OwnPermissions = append(node.OwnPermissions, perm)
		}
	}

	// Third pass: Build tree structure and calculate inherited/effective permissions
	rootNodes := []model.RoleNode{}

	for roleName, node := range roleNodesMap {
		// Calculate inherited and effective permissions
		// We use a fresh visited map for each role to allow shared ancestors
		visited := make(map[string]bool)
		node.InheritedPermissions = uc.getInheritedPermissions(roleName, parentMap, roleNodesMap, visited)
		node.EffectivePermissions = uc.mergePermissions(node.OwnPermissions, node.InheritedPermissions)

		// If this role has no parent, it's a root node
		if node.ParentID == nil {
			rootNodes = append(rootNodes, *node)
		} else {
			// Add as child to all its direct parents to build full graph visibility
			for _, pName := range node.Parents {
				if parent, exists := roleNodesMap[pName]; exists {
					parent.Children = append(parent.Children, *node)
				}
			}
		}
	}

	return &model.InheritanceTreeResponse{
		Roles: rootNodes,
	}, nil
}

// getInheritedPermissions recursively collects permissions from parent roles with cycle detection
func (uc *PermissionUseCase) getInheritedPermissions(
	roleName string,
	parentMap map[string][]string,
	roleNodesMap map[string]*model.RoleNode,
	visited map[string]bool,
) [][]string {
	if visited[roleName] {
		return [][]string{}
	}
	visited[roleName] = true

	inherited := [][]string{}

	// Get parents
	parents, hasParents := parentMap[roleName]
	if !hasParents {
		return inherited
	}

	for _, parentName := range parents {
		// Get parent node
		parentNode, exists := roleNodesMap[parentName]
		if !exists {
			continue
		}

		// Add parent's own permissions
		inherited = append(inherited, parentNode.OwnPermissions...)

		// Recursively add grandparent permissions
		// Clone visited map for each branch to correctly handle DAGs but still block cycles
		branchVisited := make(map[string]bool)
		for k, v := range visited {
			branchVisited[k] = v
		}
		grandparentPerms := uc.getInheritedPermissions(parentName, parentMap, roleNodesMap, branchVisited)
		inherited = append(inherited, grandparentPerms...)
	}

	return inherited
}

// mergePermissions combines own and inherited permissions, removing duplicates
func (uc *PermissionUseCase) mergePermissions(own, inherited [][]string) [][]string {
	permMap := make(map[string][]string)

	// Helper to create a unique key for a policy: domain|path|method
	// We omit role (perm[0]) to allow merging same permission from different roles
	getPermKey := func(p []string) string {
		if len(p) < 4 {
			return ""
		}
		return p[1] + "|" + p[2] + "|" + p[3]
	}

	// Add own permissions
	for _, perm := range own {
		key := getPermKey(perm)
		if key != "" {
			permMap[key] = perm
		}
	}

	// Add inherited permissions (won't override own)
	for _, perm := range inherited {
		key := getPermKey(perm)
		if key != "" {
			if _, exists := permMap[key]; !exists {
				permMap[key] = perm
			}
		}
	}

	// Convert back to slice
	result := make([][]string, 0, len(permMap))
	for _, perm := range permMap {
		result = append(result, perm)
	}

	return result
}
