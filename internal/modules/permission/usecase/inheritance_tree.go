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
	parentMap := make(map[string]string) // child -> parent

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
		}

		// Get parent roles (role inheritance via Casbin grouping)
		parents, err := uc.GetParentRoles(ctx, role.Name)
		if err == nil && len(parents) > 0 {
			// For simplicity, we take the first parent
			// In a real system, you might want to handle multiple inheritance
			parentMap[role.Name] = parents[0]
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
			// Check if this is a direct permission (not inherited)
			// For now, we consider all permissions as "own" and will calculate inherited later
			node.OwnPermissions = append(node.OwnPermissions, perm)
		}
	}

	// Third pass: Build tree structure and calculate inherited/effective permissions
	rootNodes := []model.RoleNode{}

	for roleName, node := range roleNodesMap {
		// Calculate inherited and effective permissions
		node.InheritedPermissions = uc.getInheritedPermissions(roleName, parentMap, roleNodesMap)
		node.EffectivePermissions = uc.mergePermissions(node.OwnPermissions, node.InheritedPermissions)

		// If this role has no parent, it's a root node
		if node.ParentID == nil {
			rootNodes = append(rootNodes, *node)
		} else {
			// Add as child to parent
			if parent, exists := roleNodesMap[*node.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	return &model.InheritanceTreeResponse{
		Roles: rootNodes,
	}, nil
}

// getInheritedPermissions recursively collects permissions from parent roles
func (uc *PermissionUseCase) getInheritedPermissions(
	roleName string,
	parentMap map[string]string,
	roleNodesMap map[string]*model.RoleNode,
) [][]string {
	inherited := [][]string{}

	// Get parent
	parentName, hasParent := parentMap[roleName]
	if !hasParent {
		return inherited
	}

	// Get parent node
	parentNode, exists := roleNodesMap[parentName]
	if !exists {
		return inherited
	}

	// Add parent's own permissions
	inherited = append(inherited, parentNode.OwnPermissions...)

	// Recursively add grandparent permissions
	grandparentPerms := uc.getInheritedPermissions(parentName, parentMap, roleNodesMap)
	inherited = append(inherited, grandparentPerms...)

	return inherited
}

// mergePermissions combines own and inherited permissions, removing duplicates
func (uc *PermissionUseCase) mergePermissions(own, inherited [][]string) [][]string {
	permMap := make(map[string][]string)

	// Add own permissions
	for _, perm := range own {
		if len(perm) < 4 {
			continue
		}
		key := perm[0] + "|" + perm[1] + "|" + perm[2] + "|" + perm[3]
		permMap[key] = perm
	}

	// Add inherited permissions (won't override own)
	for _, perm := range inherited {
		if len(perm) < 4 {
			continue
		}
		key := perm[0] + "|" + perm[1] + "|" + perm[2] + "|" + perm[3]
		if _, exists := permMap[key]; !exists {
			permMap[key] = perm
		}
	}

	// Convert back to slice
	result := make([][]string, 0, len(permMap))
	for _, perm := range permMap {
		result = append(result, perm)
	}

	return result
}
