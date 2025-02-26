package v11

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/event"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type OneBotV11Adapter struct {
	adapter.BaseAdapter
}

func NewOneBotV11Adapter(driver driver.Driver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		BaseAdapter: adapter.BaseAdapter{
			Driver: driver,
		},
	}
}

func (a *OneBotV11Adapter) Run(ctx context.Context) error {
	err := a.Driver.Connect(ctx)
	if err != nil {
		log.Error("OneBotV11Adapter.Run", err)
		return err
	}

	a.Driver.SetReceiveHandler(a.handleMessage)
	return nil
}

func (a *OneBotV11Adapter) handleMessage(d driver.Driver, msg []byte) {
	log.Debug("OneBotV11Adapter: Received message: %s", msg)
	go func(m []byte) {
		if err := a.processMessage(d, m); err != nil {
			log.Error("OneBotV11Adapter: Failed to process message: %s", err)
		}
	}(msg)
}

func (a *OneBotV11Adapter) processMessage(d driver.Driver, msg []byte) error {
	e, err := event.JsonMsgToEvent(msg)
	if err != nil {
		return err
	}

	// init bot
	if a.Bot == nil {
		lce, ok := e.(*event.MetaEventLifeCycle)
		if !ok {
			return nil
		}

		dt := d.GetDriverType()
		if (dt == driver.DriverTypeWebSocketReverse || dt == driver.DriverTypeWebSocket) && lce.SubType != event.LifeCycleSubTypeConnect {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for ws driver", lce.SubType)
			return nil
		}
		if dt == driver.DriverTypeHTTPPost && (lce.SubType == event.LifeCycleSubTypeEnable || lce.SubType == event.LifeCycleSubTypeDisable) {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for http post driver", lce.SubType)
			return nil
		}
		a.Bot = bot.NewBot(lce.SelfID)
		log.Info("OneBotV11Adapter: Bot initialized, self ID: %s", lce.SelfID)
		return nil
	}

	return nil
}
