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

package hedera_mirror_client

import (
	"github.com/hashgraph/hedera-sdk-go"
	hedera2 "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/message"
	"github.com/stretchr/testify/mock"
)

type MockHederaMirrorClient struct {
	mock.Mock
}

func (m *MockHederaMirrorClient) GetHederaTopicMessagesAfterTimestamp(topicId hedera.TopicID, timestamp int64) (*message.HederaMessages, error) {
	args := m.Called(topicId, timestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*message.HederaMessages), nil
	}
	return args.Get(0).(*message.HederaMessages), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetSuccessfulAccountCreditTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp int64) (*hedera2.Transactions, error) {
	args := m.Called(accountId, milestoneTimestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*hedera2.Transactions), nil
	}
	return args.Get(0).(*hedera2.Transactions), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountTransaction(transactionID string) (*hedera2.Transactions, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).(*hedera2.Transactions), nil
	}
	return args.Get(0).(*hedera2.Transactions), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetStateProof(transactionID string) ([]byte, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), args.Get(1).(error)

}

func (m *MockHederaMirrorClient) AccountExists(accountID hedera.AccountID) bool {
	args := m.Called(accountID)
	return args.Get(0).(bool)
}
