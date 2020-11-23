package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type ConsensusMessageHandler struct {
	repository repositories.MessageRepository
}

func NewConsensusMessageHandler(repository message.MessageRepository) *ConsensusMessageHandler {
	return &ConsensusMessageHandler{
		repository: repository,
	}
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) {
	err := cmh.handlePayload(payload)
	if err != nil {
		log.Fatalf("Error - could not handle payload: [%s]", err)
	}
}

func (cmh ConsensusMessageHandler) handlePayload(payload []byte) error {
	m := &validatorproto.TopicSignatureMessage{}
	err := proto.Unmarshal(payload, m)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal topic signature message. - [%s]", err))
	}

	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	decodedSig, err := hex.DecodeString(m.GetSignature())
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode signature. - [%s]", err))
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("%s-%s-%d-%s", ctm.TransactionId, ctm.EthAddress, ctm.Amount, ctm.Fee)))
	messageHash := hex.EncodeToString(hash)

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to recover public key. - [%s]", err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal public key. - [%s]", err))
	}

	address := crypto.PubkeyToAddress(*pubKey)
	addresses := config.LoadConfig().Hedera.Handler.ConsensusMessage.Addresses

	if !isValidAddress(address.String(), addresses) {
		return errors.New(fmt.Sprintf("Address is not valid - [%s]", address.String()))
	}

	messages, err := cmh.repository.Get(m.TransactionId, m.Signature, messageHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}

	if len(messages) > 0 {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	err = cmh.repository.Add(&message.TransactionMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
		Signature:     m.Signature,
		Hash:          messageHash,
		Leader:        false,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	txSignatures, err := cmh.repository.GetByTransactionId(m.TransactionId, messageHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	if len(txSignatures) == 1 {
		err := cmh.repository.Elect(messageHash, m.Signature)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not soft elect leader for Transaction [%s]", m.TransactionId))
		}
	}

	requiredSigCount := len(addresses)/2 + len(addresses)%2
	if len(txSignatures) > requiredSigCount {
		// Send Tx Message
	}

	return nil
}

func isValidAddress(address string, addresses []string) bool {
	for _, a := range addresses {
		if a == address {
			return true
		}
	}
	return false
}
