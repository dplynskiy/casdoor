// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"fmt"
	"time"

	PTAFLTypes "github.com/casdoor/casdoor/pt_af_logic/types"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/builder"
	"github.com/xorm-io/core"
)

const defaultStatus = string(PTAFLTypes.SubscriptionNew)

type Subscription struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Discount    int    `json:"discount"`

	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Duration    int       `json:"duration"`
	Description string    `xorm:"varchar(100)" json:"description"`
	Comment     string    `xorm:"varchar(100)" json:"comment"`

	User string `xorm:"mediumtext" json:"user"`
	Plan string `xorm:"varchar(100)" json:"plan"`

	IsEnabled   bool   `json:"isEnabled"`
	Submitter   string `xorm:"varchar(100)" json:"submitter"`
	Approver    string `xorm:"varchar(100)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`

	State string `xorm:"varchar(100)" json:"state"`

	WasPilot        bool       `json:"wasPilot"`
	PilotExpiryDate *time.Time `json:"pilotExpiryDate"`
}

func NewSubscription(owner string, user string, plan string, duration int) *Subscription {
	id := util.GenerateId()[:6]
	return &Subscription{
		Name:        "Subscription_" + id,
		DisplayName: "New Subscription - " + id,
		Owner:       owner,
		User:        owner + "/" + user,
		Plan:        owner + "/" + plan,
		CreatedTime: util.GetCurrentTime(),
		State:       defaultStatus,
		Duration:    duration,
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, duration),
	}
}

func GetSubscriptionCount(owner, field, value string, filter builder.Cond) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	if filter != nil {
		session = session.Where(filter)
	}
	return session.Count(&Subscription{})
}

func GetSubscriptions(owner string, filter builder.Cond) ([]*Subscription, error) {
	subscriptions := []*Subscription{}
	session := GetSession(owner, -1, -1, "", "", "", "")
	if filter != nil {
		session = session.Where(filter)
	}

	err := session.Find(&subscriptions)
	if err != nil {
		return subscriptions, err
	}

	return subscriptions, nil
}

func GetPaginationSubscriptions(owner string, offset, limit int, field, value, sortField, sortOrder string, filter builder.Cond) ([]*Subscription, error) {
	subscriptions := []*Subscription{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	if filter != nil {
		session = session.Where(filter)
	}
	err := session.Find(&subscriptions)
	if err != nil {
		return subscriptions, err
	}

	return subscriptions, nil
}

func getSubscription(owner string, name string, filter builder.Cond) (*Subscription, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	subscription := Subscription{Owner: owner, Name: name}
	session := adapter.Engine.NewSession()
	if filter != nil {
		session = session.Where(filter)
	}
	existed, err := session.Get(&subscription)
	if err != nil {
		return nil, err
	}

	if existed {
		return &subscription, nil
	} else {
		return nil, nil
	}
}

func GetSubscription(id string, filter builder.Cond) (*Subscription, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getSubscription(owner, name, filter)
}

func UpdateSubscription(id string, subscription *Subscription) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if s, err := getSubscription(owner, name, nil); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(subscription)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddSubscription(subscription *Subscription) (bool, error) {
	affected, err := adapter.Engine.Insert(subscription)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteSubscription(subscription *Subscription) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{subscription.Owner, subscription.Name}).Delete(&Subscription{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (subscription *Subscription) GetId() string {
	return fmt.Sprintf("%s/%s", subscription.Owner, subscription.Name)
}
