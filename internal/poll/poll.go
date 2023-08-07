// Copyright (C) 2023 Patrice Congo <@congop>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package poll

import (
	"context"
	"reflect"
	"time"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Poller[Out any] struct {
	LastOutcome          Out
	Interval             time.Duration
	Timeout              time.Duration
	MaxConsecutiveErrors int

	OutcomeNillable bool
	OutcomeGetter   func() (Out, error)
	ConditionFunc   func(outcome Out) (bool, error)

	consecutiveErrorCount uint
	zeroOut               Out
}

func (p Poller[Out]) maxConsecutiveErrorsReqMet() bool {
	return p.MaxConsecutiveErrors < 0 || uint(p.MaxConsecutiveErrors) > p.consecutiveErrorCount
}

func (poller *Poller[Out]) Poll(ctx context.Context) error {
	untilTaskStopFunc := func(ctx context.Context) (bool, error) {
		outcome, err := poller.OutcomeGetter()
		poller.LastOutcome = outcome
		if err != nil {
			log.Tracef(ctx, "Poller.poll -- outcomeGetter error(%T):%s", err, err.Error())
			poller.consecutiveErrorCount++
			if !poller.maxConsecutiveErrorsReqMet() {
				return true, errors.Errorf(
					"Poller.poll -- max error exceeded: count=%d, max=%d err(%T)=%s",
					poller.consecutiveErrorCount, poller.MaxConsecutiveErrors, err, err.Error())
			}
			return false, nil
		}
		// outcome == nil && !poller.OutcomeNillable
		if !poller.OutcomeNillable && reflect.DeepEqual(outcome, poller.zeroOut) {
			poller.consecutiveErrorCount++
			if poller.maxConsecutiveErrorsReqMet() {
				return true, errors.Errorf(
					"Poller.poll -- max error exceeded: count=%d, max=%d err=%s",
					poller.consecutiveErrorCount, poller.MaxConsecutiveErrors, "No outcome available")
			}
			return false, nil
		}

		poller.consecutiveErrorCount = 0

		stop, err := poller.ConditionFunc(outcome)
		// TODO check type error
		return stop, err
	}

	return wait.PollUntilContextTimeout(ctx, poller.Interval, poller.Timeout, true, untilTaskStopFunc)
}
