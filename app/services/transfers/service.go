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

package transfers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger             *log.Entry
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	fees               service.Fees
	ethSigner          service.Signer
	transferRepository repository.Transfer
	topicID            hedera.TopicID
}

func NewService(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	fees service.Fees,
	signer service.Signer,
	transferRepository repository.Transfer,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		logger:             config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		fees:               fees,
		ethSigner:          signer,
		transferRepository: transferRepository,
		topicID:            tID,
	}
}

// SanityCheck performs validation on the memo and state proof for the transaction
func (ts *Service) SanityCheckTransfer(tx mirror_node.Transaction) (*memo.Memo, error) {
	m, e := memo.FromBase64String(tx.MemoBase64)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse transaction memo. Error: [%s]", e))
	}

	// TODO: Uncomment when State Proof Method gets updated accordingly
	//stateProof, e := ts.mirrorNode.GetStateProof(tx.TransactionID)
	//if e != nil {
	//	return nil, errors.New(fmt.Sprintf("Could not GET state proof. Error [%s]", e))
	//}
	//
	//verified, e := proof.Verify(tx.TransactionID, stateProof)
	//if e != nil {
	//	return nil, errors.New(fmt.Sprintf("State proof verification failed. Error [%s]", e))
	//}
	//
	//if !verified {
	//	return nil, errors.New("State proof not valid")
	//}

	return m, nil
}

// InitiateNewTransfer Stores the incoming transfer message into the Database aware of already processed transfers
func (ts *Service) InitiateNewTransfer(tm encoding.TransferMessage) (*transfer.Transfer, error) {
	dbTransaction, err := ts.transferRepository.GetByTransactionId(tm.TransactionId)
	if err != nil {
		ts.logger.Errorf("Failed to get record with TransactionID [%s]. Error [%s]", tm.TransactionId, err)
		return nil, err
	}

	if dbTransaction != nil {
		ts.logger.Infof("Transaction with ID [%s] already added", tm.TransactionId)
		return dbTransaction, err
	}

	ts.logger.Debugf("Adding new Transaction Record TX ID [%s]", tm.TransactionId)
	tx, err := ts.transferRepository.Create(tm.TransferMessage)
	if err != nil {
		ts.logger.Errorf("Failed to create a transaction record for TransactionID [%s]. Error [%s].", tm.TransactionId, err)
		return nil, err
	}
	return tx, nil
}

// SaveRecoveredTxn creates new Transaction record persisting the recovered Transfer TXn
func (ts *Service) SaveRecoveredTxn(txId, amount, sourceAsset, targetAsset string, m memo.Memo) error {
	err := ts.transferRepository.SaveRecoveredTxn(&validatorproto.TransferMessage{
		TransactionId:         txId,
		Receiver:              m.EthereumAddress,
		Amount:                amount,
		TxReimbursement:       m.TxReimbursementFee,
		GasPrice:              m.GasPrice,
		SourceAsset:           sourceAsset,
		TargetAsset:           targetAsset,
		ExecuteEthTransaction: m.ExecuteEthTransaction,
	})
	if err != nil {
		ts.logger.Errorf("Something went wrong while saving new Recovered Transaction with ID [%s]. Error [%s]", txId, err)
		return err
	}

	ts.logger.Infof("Added new Transaction Record with Txn ID [%s]", txId)
	return err
}

// VerifyFee verifies that the provided TX reimbursement fee is enough using the
// Fee Calculator and updates the Transaction Record to Insufficient Fee if necessary
func (ts *Service) VerifyFee(tm encoding.TransferMessage) error {
	isSufficient, err := ts.fees.ValidateExecutionFee(tm.TxReimbursement, tm.Amount, tm.GasPrice)
	if !isSufficient {
		ts.logger.Errorf("Fee validation for TX ID [%s] failed. Provided tx reimbursement fee is invalid/insufficient. Error [%s].", tm.TransactionId, err)
		if err := ts.transferRepository.UpdateStatusInsufficientFee(tm.TransactionId); err != nil {
			ts.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transfer.StatusInsufficientFee, tm.TransactionId, err)
			return err
		}

		ts.logger.Debugf("TX with ID [%s] was updated to [%s]. Provided TxReimbursement [%s].", tm.TransactionId, transfer.StatusInsufficientFee, tm.TxReimbursement)
		return err
	}
	return nil
}

func (ts *Service) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ts.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureMined(txId)
		if err != nil {
			ts.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onRevert = func() {
		ts.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			ts.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ts *Service) ProcessTransfer(tm encoding.TransferMessage) error {
	gasPriceWeiBn, err := helper.ToBigInt(tm.GasPrice)
	if err != nil {
		ts.logger.Errorf("Failed to parse Gas Price Wei for TX ID [%s] to a big integer [%s]. Error [%s].", tm.TransactionId, tm.GasPrice, err)
		return err
	}

	authMsgHash, err := auth_message.EncodeBytesFrom(tm.TransactionId, tm.Receiver, tm.TargetAsset, tm.Amount, tm.TxReimbursement, gasPriceWeiBn.String())
	if err != nil {
		ts.logger.Errorf("Failed to encode the authorisation signature for TX ID [%s]. Error: %s", tm.TransactionId, err)
		return err
	}

	signatureBytes, err := ts.ethSigner.Sign(authMsgHash)
	if err != nil {
		ts.logger.Errorf("Failed to sign the authorisation signature for TX ID [%s]. Error: %s", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)

	signatureMessage := encoding.NewSignatureMessage(
		tm.TransactionId,
		tm.Receiver,
		tm.Amount,
		tm.TxReimbursement,
		tm.GasPrice,
		signature,
		tm.TargetAsset)

	tsm := signatureMessage.GetTopicSignatureMessage()
	sigMsgBytes, err := signatureMessage.ToBytes()
	if err != nil {
		ts.logger.Errorf("Failed to encode Signature Message to bytes for TX [%s]. Error %s", err, tsm.TransferID)
		return err
	}

	messageTxId, err := ts.hederaNode.SubmitTopicConsensusMessage(
		ts.topicID,
		sigMsgBytes)
	if err != nil {
		ts.logger.Errorf("Failed to submit Signature Message to Topic for TX [%s]. Error: %s", tsm.TransferID, err)
		return err
	}

	// Update Transfer Record
	tx, err := ts.transferRepository.GetByTransactionId(tsm.TransferID)
	if err != nil {
		ts.logger.Errorf("Failed to get TX [%s] from DB", tsm.TransferID)
		return err
	}

	tx.Status = transfer.StatusInProgress
	tx.SignatureMsgStatus = transfer.StatusSignatureSubmitted
	err = ts.transferRepository.Save(tx)
	if err != nil {
		ts.logger.Errorf("Failed to update TX [%s]. Error [%s].", tsm.TransferID, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	ts.logger.Infof("Submitted signature for TX ID [%s] on Topic [%s]", tsm.TransferID, ts.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := ts.authMessageSubmissionCallbacks(tsm.TransferID)
	ts.mirrorNode.WaitForTransaction(messageTxId.String(), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}
