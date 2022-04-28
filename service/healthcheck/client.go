/**
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

package healthcheck

import (
	"context"
	"time"

	api "github.com/polarismesh/polaris-server/common/api/v1"
	"github.com/polarismesh/polaris-server/plugin"
)

const clientPrefix = "c_"

func toClientId(instanceId string) string {
	return clientPrefix + instanceId
}

func (s *Server) doReportByClient(ctx context.Context, client *api.Client) *api.Response {
	if len(s.checkers) == 0 {
		return api.NewResponse(api.HealthCheckNotOpen)
	}
	checker, ok := s.checkers[int32(api.HealthCheck_HEARTBEAT)]
	if !ok {
		return api.NewClientResponse(api.HeartbeatTypeNotFound, client)
	}
	request := &plugin.ReportRequest{
		QueryRequest: plugin.QueryRequest{
			InstanceId: toClientId(client.GetId().GetValue()),
			Host:       client.GetHost().GetValue(),
		},
		LocalHost:  s.localHost,
		CurTimeSec: time.Now().Unix() - s.timeAdjuster.GetDiff(),
	}
	err := checker.Report(request)
	if err != nil {
		log.Errorf("[Heartbeat][Server]fail to do report client for %s, id is %s, err is %v",
			client.GetHost().GetValue(), client.GetId().GetValue(), err)
		return api.NewClientResponse(api.HeartbeatException, client)
	}
	return api.NewClientResponse(api.ExecuteSuccess, client)
}
