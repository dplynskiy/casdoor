package jobs

import (
	"fmt"
	"time"

	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/pt_af_logic"
	PTAFLTypes "github.com/casdoor/casdoor/pt_af_logic/types"
	"github.com/xorm-io/builder"
)

const pilotExpiringLogPrefix = "PilotExpiring Job: "

type pilotExpiring struct {
}

// NewPilotExpiring конструктор объекта для запуска задания по переводу просроченных пилотов в статус Истек срок Пилота
func NewPilotExpiring() *pilotExpiring {
	return &pilotExpiring{}
}

// Run ...
func (j *pilotExpiring) Run() {
	mskLoc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		logs.Error(fmt.Errorf(pilotExpiringLogPrefix+"time.LoadLocation: %w", err).Error())
		return
	}

	filter := builder.Eq{
		"state": PTAFLTypes.SubscriptionPilot.String(),
	}.And(builder.Lte{
		"pilot_expiry_date": time.Now().In(mskLoc),
	})

	expiredPilotSubscriptions, err := object.GetSubscriptions("", filter)
	if err != nil {
		logs.Error(fmt.Errorf(pilotExpiringLogPrefix+"object.GetSubscriptions: %w", err).Error())
		return
	}

	logs.Info(fmt.Errorf(pilotExpiringLogPrefix+"got %v expired subscriptions ", len(expiredPilotSubscriptions)).Error())
	if len(expiredPilotSubscriptions) == 0 {
		return
	}

	ctx := createCtx()
	systemUser, err := getSystemUser()
	if err != nil {
		logs.Error(fmt.Errorf(pilotExpiringLogPrefix+"getSystemUser: %w", err).Error())
		return
	}

	for _, expiredPilotSubscription := range expiredPilotSubscriptions {
		oldSubscription := *expiredPilotSubscription
		expiredPilotSubscription.State = PTAFLTypes.SubscriptionPilotExpired.String()
		affected, err := pt_af_logic.ProcessSubscriptionUpdate(ctx, systemUser, expiredPilotSubscription, &oldSubscription)
		if err != nil {
			logs.Error(fmt.Errorf(pilotExpiringLogPrefix+"pt_af_logic.ProcessSubscriptionUpdate: %w", err).Error())
			continue
		}
		if !affected {
			logs.Error(fmt.Errorf(pilotExpiringLogPrefix+"ProcessSubscriptionUpdate not affected for id: %s", expiredPilotSubscription.GetId()).Error())
		}
	}
}
