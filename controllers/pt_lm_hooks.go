//Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package controllers

import (
	"encoding/json"
	"net/url"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/pt_af_logic"
	"github.com/casdoor/casdoor/util"
)

// UpdateSubscriptionPostBack ...
// @Title Blueprint
// @Tag PTLMHOOKS API
// @Description PTLMHOOKS
// @Param   body    body   object.Record  true        "The details of the event"
// @Success 200 {object} controllers.Response The Response object
// @router /update-subscription-postback [post]
func (c *ApiController) UpdateSubscriptionPostBack() {

	var record object.Record
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &record)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if record.Action != "update-subscription" {
		return
	}

	u, err := url.Parse(record.RequestUri)
	if err != nil {
		panic(err)
	}
	id := u.Query().Get("id")

	subscription, _ := object.GetSubscription(id)
	if subscription == nil {
		util.LogWarning(c.Ctx, "No subscription found")
		c.ServeJSON() // to avoid crash
		return
	}

	switch object.SubscriptionState(subscription.State) {
	case object.SubscriptionStarted:
		{
			err := pt_af_logic.CreateTenant(c.Ctx, subscription)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}
	}

	c.ServeJSON()
}
