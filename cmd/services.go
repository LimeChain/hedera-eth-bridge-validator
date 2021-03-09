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

package main

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/contracts"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

type ServicesContext struct {
	ethSigner service.Signer
	scheduler service.Scheduler
	contracts service.Contracts
	bridge    service.Bridge
	fees      service.Fees
}

// PrepareServices instantiates all the necessary services with their required context and parameters
func PrepareServices(c config.Config, clients client.Clients, repositories Repositories) *ServicesContext {
	ethSigner := eth.NewEthSigner(c.Hedera.Client.Operator.EthPrivateKey)
	contractService := contracts.NewService(clients.Ethereum, c.Hedera.Eth)
	schedulerService := scheduler.NewScheduler(c.Hedera.Handler.ConsensusMessage.TopicId, ethSigner.Address(),
		c.Hedera.Handler.ConsensusMessage.SendDeadline, contractService, clients.HederaNode)
	feeService := fees.NewCalculator(clients.ExchangeRate, c.Hedera, contractService)
	bridgeService := bridge.NewService(
		clients,
		repositories.transaction,
		repositories.message,
		contractService,
		feeService,
		ethSigner,
		c.Hedera.Watcher.ConsensusMessage.Topic.Id)

	return &ServicesContext{
		ethSigner: ethSigner,
		scheduler: schedulerService,
		contracts: contractService,
		bridge:    bridgeService,
		fees:      feeService,
	}
}