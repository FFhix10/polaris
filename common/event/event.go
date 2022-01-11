/*
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package event

import (
	"github.com/polarismesh/polaris-server/common/log"
	"go.uber.org/zap"
	"sync"
)

const (
	QueueSize = 10240
)

// Event 事件对象，包含类型和事件消息
type Event struct {
	EventType string
	Message   interface{}
}

type Callback func(event Event) bool

// Center 事件中心
type Center struct {
	watchers *sync.Map
	lock     *sync.Mutex
	queue    chan Event
}

// NewEventCenter 新建事件中心
func NewEventCenter() *Center {
	center := &Center{
		watchers: new(sync.Map),
		lock:     new(sync.Mutex),
		queue:    make(chan Event, QueueSize),
	}

	go func() {
		center.handlerEvent()
	}()

	return center
}

// FireEvent 发布一个事件
func (c *Center) FireEvent(event Event) {
	log.GetDefaultLogger().Infof("[Common][Event] fire event.", event.EventType)

	select {
	case c.queue <- event:
		return
	}
}

// WatchEvent 监听事件
func (c *Center) WatchEvent(eventType string, cb Callback) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cbs, ok := c.watchers.Load(eventType)
	if !ok {
		cbs = []Callback{cb}
		c.watchers.Store(eventType, cbs)
	} else {
		cbArr := cbs.([]Callback)
		cbArr = append(cbArr, cb)
	}
}

func (c *Center) handlerEvent() {
	defer func() {
		if err := recover(); err != nil {
			log.GetDefaultLogger().Error("[Common][Event] handler event error.", zap.Any("error", err))
		}
	}()

	for e := range c.queue {
		cbs, ok := c.watchers.Load(e.EventType)
		if !ok {
			continue
		}

		cbArr := cbs.([]Callback)
		for _, cb := range cbArr {
			ok := cb(e)
			if !ok {
				log.GetDefaultLogger().Errorf("[Common][Event] cb message error. event = %+v", e)
			}
		}
	}
}
