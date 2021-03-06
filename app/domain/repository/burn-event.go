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

package repository

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"

type BurnEvent interface {
	Create(id string, amount int64, recipient string) error
	UpdateStatusSubmitted(id, scheduleID, transactionId string) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusFailed(txId string) error
	// Returns BurnEvent by its Id (represented in {ethTxHash}-{logIndex})
	Get(txId string) (*entity.BurnEvent, error)
}
