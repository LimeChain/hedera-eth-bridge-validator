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

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
)

type Transfer interface {
	GetByTransactionId(transactionId string) (*transfer.Transfer, error)
	GetInitialAndSignatureSubmittedTx() ([]*transfer.Transfer, error)
	GetUnprocessedTransfers() ([]transfer.Transfer, error)

	Create(ct *proto.TransferMessage) (*transfer.Transfer, error)
	Save(tx *transfer.Transfer) error
	SaveRecoveredTxn(ct *proto.TransferMessage) error
	UpdateStatusInsufficientFee(txId string) error
	UpdateStatusCompleted(txId string) error

	UpdateStatusSignatureMined(txId string) error
	UpdateStatusSignatureFailed(txId string) error

	UpdateEthTxSubmitted(txId string, hash string) error
	UpdateEthTxMined(txId string) error
	UpdateEthTxReverted(txId string) error

	UpdateStatusEthTxMsgSubmitted(txId string) error
	UpdateStatusEthTxMsgMined(txId string) error
	UpdateStatusEthTxMsgFailed(txId string) error
}
