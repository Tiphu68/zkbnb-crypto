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

package std

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/std/algebra/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"log"
	"zecrey-crypto/hash/bn254/zmimc"
	"zecrey-crypto/zecrey/twistededwards/tebn254/zecrey"
)

type UnlockProofConstraints struct {
	// A
	A_pk Point
	// response
	Z_sk, Z_skInv Variable
	// common inputs
	Pk          Point
	ChainId     Variable
	AssetId     Variable
	Balance     Variable
	DeltaAmount Variable
	// gas fee
	A_T_feeC_feeRPrimeInv Point
	Z_bar_r_fee           Variable
	C_fee                 ElGamalEncConstraints
	T_fee                 Point
	GasFeeAssetId         Variable
	GasFee                Variable
	IsEnabled             Variable
}

// define for testing transfer proof
func (circuit UnlockProofConstraints) Define(curveID ecc.ID, api API) error {
	// first check if C = c_1 \oplus c_2
	// get edwards curve params
	params, err := twistededwards.NewEdCurve(curveID)
	if err != nil {
		return err
	}
	// mimc
	hFunc, err := mimc.NewMiMC(zmimc.SEED, curveID, api)
	if err != nil {
		return err
	}
	H := Point{
		X: api.Constant(HX),
		Y: api.Constant(HY),
	}
	VerifyUnlockProof(api, circuit, params, hFunc, H)
	return nil
}

func VerifyUnlockProof(
	api API,
	proof UnlockProofConstraints,
	params twistededwards.EdCurve,
	hFunc MiMC,
	h Point,
) {
	tool := NewEccTool(api, params)
	hFunc.Write(FixedCurveParam(api))
	writePointIntoBuf(&hFunc, proof.Pk)
	writePointIntoBuf(&hFunc, proof.A_pk)
	hFunc.Write(proof.ChainId)
	hFunc.Write(proof.AssetId)
	hFunc.Write(proof.Balance)
	hFunc.Write(proof.DeltaAmount)
	// gas fee
	writePointIntoBuf(&hFunc, proof.A_T_feeC_feeRPrimeInv)
	writeEncIntoBuf(&hFunc, proof.C_fee)
	hFunc.Write(proof.GasFeeAssetId)
	hFunc.Write(proof.GasFee)
	c := hFunc.Sum()
	var l, r Point
	l.ScalarMulFixedBase(api, params.BaseX, params.BaseY, proof.Z_sk, params)
	r.ScalarMulNonFixedBase(api, &proof.Pk, c, params)
	r.AddGeneric(api, &r, &proof.A_pk, params)
	IsPointEqual(api, proof.IsEnabled, l, r)
	// check gas fee proof
	var (
		hNeg                                               Point
		C_feePrimeInv, C_feeLprimeInv, T_feeDivC_feeRprime Point
	)
	hNeg.Neg(api, &h)
	C_feeDelta := tool.ScalarMul(hNeg, proof.GasFee)
	C_feeLprimeInv.Neg(api, &proof.C_fee.CL)
	C_feeLprimeInv.Neg(api, &proof.C_fee.CL)
	C_feePrimeInv = tool.AddPoint(proof.C_fee.CR, C_feeDelta)
	C_feePrimeInv.Neg(api, &C_feePrimeInv)
	T_feeDivC_feeRprime = tool.AddPoint(proof.T_fee, C_feePrimeInv)
	// Verify T(C_R - C_R^{\star})^{-1} = (C_L - C_L^{\star})^{-sk^{-1}} g^{\bar{r}}
	l2 := tool.AddPoint(tool.ScalarBaseMul(proof.Z_bar_r_fee), tool.ScalarMul(C_feeLprimeInv, proof.Z_skInv))
	r2 := tool.AddPoint(proof.A_T_feeC_feeRPrimeInv, tool.ScalarMul(T_feeDivC_feeRprime, c))
	IsPointEqual(api, proof.IsEnabled, l2, r2)
}

func SetEmptyUnlockProofWitness() (witness UnlockProofConstraints) {
	witness.A_pk, _ = SetPointWitness(BasePoint)
	witness.Pk, _ = SetPointWitness(BasePoint)
	// response
	witness.Z_sk.Assign(ZeroInt)
	witness.Z_skInv.Assign(ZeroInt)
	witness.ChainId.Assign(ZeroInt)
	witness.AssetId.Assign(ZeroInt)
	witness.Balance.Assign(ZeroInt)
	witness.DeltaAmount.Assign(ZeroInt)
	// gas fee
	witness.A_T_feeC_feeRPrimeInv, _ = SetPointWitness(BasePoint)
	witness.Z_bar_r_fee.Assign(ZeroInt)
	witness.C_fee, _ = SetElGamalEncWitness(ZeroElgamalEnc)
	witness.T_fee, _ = SetPointWitness(BasePoint)
	witness.GasFeeAssetId.Assign(ZeroInt)
	witness.GasFee.Assign(ZeroInt)
	witness.IsEnabled = SetBoolWitness(false)
	return witness
}

// set the witness for RemoveLiquidity proof
func SetUnlockProofWitness(proof *zecrey.UnlockProof, isEnabled bool) (witness UnlockProofConstraints, err error) {
	if proof == nil {
		log.Println("[SetUnlockProofWitness] invalid params")
		return witness, err
	}

	// proof must be correct
	verifyRes, err := proof.Verify()
	if err != nil {
		log.Println("[SetUnlockProofWitness] invalid proof:", err)
		return witness, err
	}
	if !verifyRes {
		log.Println("[SetUnlockProofWitness] invalid proof")
		return witness, errors.New("[SetUnlockProofWitness] invalid proof")
	}

	witness.A_pk, err = SetPointWitness(proof.A_pk)
	if err != nil {
		return witness, err
	}
	witness.Pk, err = SetPointWitness(proof.Pk)
	if err != nil {
		return witness, err
	}
	// response
	witness.Z_sk.Assign(proof.Z_sk)
	witness.Z_skInv.Assign(proof.Z_skInv)
	witness.ChainId.Assign(uint64(proof.ChainId))
	witness.AssetId.Assign(uint64(proof.AssetId))
	witness.Balance.Assign(proof.Balance)
	witness.DeltaAmount.Assign(proof.DeltaAmount)
	// gas fee
	witness.A_T_feeC_feeRPrimeInv, err = SetPointWitness(proof.A_T_feeC_feeRPrimeInv)
	if err != nil {
		return witness, err
	}
	witness.Z_bar_r_fee.Assign(proof.Z_bar_r_fee)
	witness.C_fee, err = SetElGamalEncWitness(proof.C_fee)
	if err != nil {
		return witness, err
	}
	witness.T_fee, err = SetPointWitness(proof.T_fee)
	if err != nil {
		return witness, err
	}
	witness.GasFeeAssetId.Assign(uint64(proof.GasFeeAssetId))
	witness.GasFee.Assign(proof.GasFee)
	witness.IsEnabled = SetBoolWitness(isEnabled)
	return witness, nil
}
