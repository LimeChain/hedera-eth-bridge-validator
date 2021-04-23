/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"

// BurnEvent is the major service used for processing BurnEvent operations
type BurnEvent interface {
	// ProcessEvent processes the burn event by submitting the appropriate
	// scheduled transaction, leaving the synchronization of the actual transfer on HCS
	ProcessEvent(event burn_event.BurnEvent)
	// ScheduledTxID returns from the database the corresponding scheduled transaction id
	ScheduledTxID(id string) (string, error)
}
