#!/bin/bash
sed -i -e '/roleService := roleUC.NewRoleUseCase/ {
  N; N; N; N; N; N; N;
  s/roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo, permService)\n\n\taRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)\n\taccessService := accessUC.NewAccessUseCase(aRepo, env.Logger)\n\n\tuRepo := userRepo.NewUserRepository(env.DB, env.Logger)\n\tpermService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)/aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)\n\taccessService := accessUC.NewAccessUseCase(aRepo, env.Logger)\n\n\tuRepo := userRepo.NewUserRepository(env.DB, env.Logger)\n\tpermService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)\n\troleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo, permService)/
}' tests/integration/scenarios/rbac_orchestration_test.go
