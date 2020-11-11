package main

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederasdk "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	consensus_message "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	crypto_transfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	"log"
)

func main() {
	configuration := config.LoadConfig()
	persistence.RunDb(configuration.Hedera.Validator.Db)
	server := server.NewServer()
	hederaClient := hederasdk.NewClient(configuration.Hedera.MirrorNode.ApiAddress, configuration.Hedera.MirrorNode.ClientAddress)

	failure := addCryptoTransferWatchers(configuration, hederaClient, server)
	if failure != nil {
		log.Println(failure)
	}

	failure = addConsensusTopicWatchers(configuration, hederaClient, server)
	if failure != nil {
		log.Println(failure)
	}

	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func addCryptoTransferWatchers(configuration *config.Config, hederaClient *hederasdk.Client, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.CryptoTransfer.Accounts) == 0 {
		fmt.Println("There are no Crypto Transfer Watchers.")
	}
	for _, account := range configuration.Hedera.Watcher.CryptoTransfer.Accounts {
		id, e := hedera.AccountIDFromString(account.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
		}

		server.AddWatcher(crypto_transfer.NewCryptoTransferWatcher(hederaClient, id, configuration.Hedera.MirrorNode.PollingInterval))
		log.Printf("Added a Crypto Transfer Watcher for account [%s]\n", account.Id)
	}
	return nil
}

func addConsensusTopicWatchers(configuration *config.Config, hederaClient *hederasdk.Client, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.ConsensusMessage.Topics) == 0 {
		fmt.Println("There are no Consensus Topic Watchers.")
	}
	for _, topic := range configuration.Hedera.Watcher.ConsensusMessage.Topics {
		id, e := hedera.TopicIDFromString(topic.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
		}

		server.AddWatcher(consensus_message.NewConsensusTopicWatcher(hederaClient, id))
		log.Printf("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
	}
	return nil
}
