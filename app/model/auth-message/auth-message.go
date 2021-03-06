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

package auth_message

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
)

// EncodeBytesFrom returns the array of bytes representing an
// authorisation signature ready to be signed by Ethereum Private Key
func EncodeBytesFrom(txId, routerAddress, wrappedAsset, receiverEthAddress, amount string) ([]byte, error) {
	args, err := generateArguments()
	if err != nil {
		return nil, err
	}
	amountBn, err := big_numbers.ToBigInt(amount)
	if err != nil {
		return nil, err
	}

	bytesToHash, err := args.Pack([]byte(txId), common.HexToAddress(routerAddress), common.HexToAddress(wrappedAsset), common.HexToAddress(receiverEthAddress), amountBn)
	return keccak(bytesToHash), nil
}

func generateArguments() (abi.Arguments, error) {
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{
			Type: bytesType,
		},
		{
			Type: addressType,
		},
		{
			Type: addressType,
		},
		{
			Type: addressType,
		},
		{
			Type: uint256Type,
		}}, nil
}

func keccak(encodedData []byte) []byte {
	toEthSignedMsg := []byte("\x19Ethereum Signed Message:\n32")
	hash := crypto.Keccak256(encodedData)
	return crypto.Keccak256(toEthSignedMsg, hash)
}
