package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
)

// GetResourceAggregation aggregates permissions by resource with CRUD mapping
func (uc *PermissionUseCase) GetResourceAggregation(ctx context.Context) (*model.ResourceAggregationResponse, error) {
	// Use the directly available AccessRepo
	accessRights, err := uc.AccessRepo.GetAccessRights(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get access rights: %v", err)
		return nil, err
	}

	// Get all roles
	roles, err := uc.RoleRepo.FindAll(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get roles: %v", err)
		return nil, err
	}

	// Get all permissions
	allPerms, err := uc.GetAllPermissions(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get permissions: %v", err)
		return nil, err
	}

	// Build resource map
	resourceMap := make(map[string]*model.ResourcePermission)

	// Process access rights to group endpoints by resource
	for _, ar := range accessRights {
		resourceName, basePath := ar.Name, "/api/v1/"+strings.ToLower(ar.Name)
		if len(ar.Endpoints) > 0 {
			// Try to extract more accurate name from path if endpoints exist
			extractedName, extractedPath := extractResourceFromPath(ar.Endpoints[0].Path)
			resourceName = extractedName
			basePath = extractedPath
		}

		if _, exists := resourceMap[resourceName]; !exists {
			resourceMap[resourceName] = &model.ResourcePermission{
				Name:            resourceName,
				BasePath:        basePath,
				RolePermissions: make(map[string]model.ResourceCRUD),
				EndpointCount:   0,
			}
		}

		// If no endpoints, still ensure all roles have empty CRUD
		for _, role := range roles {
			if _, exists := resourceMap[resourceName].RolePermissions[role.Name]; !exists {
				resourceMap[resourceName].RolePermissions[role.Name] = model.ResourceCRUD{}
			}
		}

		if len(ar.Endpoints) == 0 {
			continue
		}

		for _, endpoint := range ar.Endpoints {
			resourceMap[resourceName].EndpointCount++

			// Map permissions for each role
			for _, role := range roles {
				crud := resourceMap[resourceName].RolePermissions[role.Name]

				// Check if this role has permission for this endpoint
				if hasPermission(allPerms, role.Name, endpoint.Path, endpoint.Method) {
					crud = mapMethodToCRUD(endpoint.Method, crud)
					resourceMap[resourceName].RolePermissions[role.Name] = crud
				}
			}
		}
	}
	// Convert map to slice
	resources := make([]model.ResourcePermission, 0, len(resourceMap))
	for _, res := range resourceMap {
		resources = append(resources, *res)
	}

	return &model.ResourceAggregationResponse{
		Resources: resources,
	}, nil
}

// extractResourceFromPath extracts resource name and base path from API path
func extractResourceFromPath(path string) (string, string) {
	// Remove /api/v1 prefix
	re := regexp.MustCompile(`^/api/v\d+/`)
	cleanPath := re.ReplaceAllString(path, "/")

	// Extract first segment
	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if len(parts) == 0 {
		return "Unknown", path
	}

	resourceName := parts[0]
	// Capitalize first letter
	if len(resourceName) > 0 {
		resourceName = strings.ToUpper(resourceName[:1]) + resourceName[1:]
	}

	// Build base path
	basePath := "/api/v1/" + parts[0]

	return resourceName, basePath
}

// mapMethodToCRUD maps HTTP method to CRUD operation
func mapMethodToCRUD(method string, current model.ResourceCRUD) model.ResourceCRUD {
	method = strings.ToUpper(method)

	switch method {
	case "GET":
		current.Read = true
	case "POST":
		current.Create = true
	case "PUT", "PATCH":
		current.Update = true
	case "DELETE":
		current.Delete = true
	}

	return current
}

// hasPermission checks if a role has permission for a specific endpoint
func hasPermission(permissions [][]string, role, path, method string) bool {
	for _, perm := range permissions {
		if len(perm) < 4 {
			continue
		}
		if perm[0] == role && perm[2] == path && perm[3] == method {
			return true
		}
	}
	return false
}
