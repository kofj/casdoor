// Copyright 2022 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"encoding/json"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// Enforce
// @Title Enforce
// @Tag Enforce API
// @Description Call Casbin Enforce API
// @Param   body    body   object.CasbinRequest  true   "Casbin request"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Param   resourceId    query   string  false   "resource id"
// @Success 200 {object} controllers.Response The Response object
// @router /enforce [post]
func (c *ApiController) Enforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")
	resourceId := c.Input().Get("resourceId")
	enforcerId := c.Input().Get("enforcerId")

	if len(c.Ctx.Input.RequestBody) == 0 {
		c.ResponseError("The request body should not be empty")
		return
	}

	var request object.CasbinRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res, err := enforcer.Enforce(request...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(res)
		return
	}

	if permissionId != "" {
		permission, err := object.GetPermission(permissionId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := []bool{}

		if permission == nil {
			res = append(res, false)
		} else {
			enforceResult, err := object.Enforce(permission, &request)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			res = append(res, enforceResult)
		}

		c.ResponseOk(res)
		return
	}

	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err = object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if resourceId != "" {
		permissions, err = object.GetPermissionsByResource(resourceId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	res := []bool{}

	listPermissionIdMap := object.GroupPermissionsByModelAdapter(permissions)
	for _, permissionIds := range listPermissionIdMap {
		firstPermission, err := object.GetPermission(permissionIds[0])
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		enforceResult, err := object.Enforce(firstPermission, &request, permissionIds...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
	}

	c.ResponseOk(res)
}

// BatchEnforce
// @Title BatchEnforce
// @Tag Enforce API
// @Description Call Casbin BatchEnforce API
// @Param   body    body   object.CasbinRequest  true   "array of casbin requests"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Success 200 {object} controllers.Response The Response object
// @router /batch-enforce [post]
func (c *ApiController) BatchEnforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")
	enforcerId := c.Input().Get("enforcerId")

	var requests []object.CasbinRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &requests)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res, err := enforcer.BatchEnforce(requests)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(res)
		return
	}

	if permissionId != "" {
		permission, err := object.GetPermission(permissionId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := [][]bool{}

		if permission == nil {
			l := len(requests)
			resRequest := make([]bool, l)
			for i := 0; i < l; i++ {
				resRequest[i] = false
			}

			res = append(res, resRequest)
		} else {
			enforceResult, err := object.BatchEnforce(permission, &requests)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			res = append(res, enforceResult)
		}

		c.ResponseOk(res)
		return
	}

	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err = object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	res := [][]bool{}

	listPermissionIdMap := object.GroupPermissionsByModelAdapter(permissions)
	for _, permissionIds := range listPermissionIdMap {
		firstPermission, err := object.GetPermission(permissionIds[0])
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		enforceResult, err := object.BatchEnforce(firstPermission, &requests, permissionIds...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
	}

	c.ResponseOk(res)
}

func (c *ApiController) GetAllObjects() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	objects, err := object.GetAllObjects(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(objects)
}

func (c *ApiController) GetAllActions() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	actions, err := object.GetAllActions(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(actions)
}

func (c *ApiController) GetAllRoles() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	roles, err := object.GetAllRoles(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(roles)
}
