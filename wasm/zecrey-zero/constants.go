/*
 * Copyright © 2021 Zecrey Protocol
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
 *
 */

package zecrey_zero

import (
	"github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"github.com/zecrey-labs/zecrey-crypto/elgamal/twistededwards/tebn254/twistedElgamal"
)

const (
	Noop = iota
	Deposit
	Lock
	Unlock
	Transfer
	Swap
	AddLiquidity
	RemoveLiquidity
	Withdraw
	DepositNft
	MintNft
	TransferNft
	SetNftPrice
	BuyNft
	WithdrawNft
)

type (
	Point      = tebn254.Point
	ElGamalEnc = twistedElgamal.ElGamalEnc
)