package bfv

import (
	"fmt"
	"testing"
)

func Benchmark_BFVScheme(b *testing.B) {

	paramSets := DefaultParams

	for _, params := range paramSets {

		bfvContext := NewBfvContext()
		if err := bfvContext.SetParameters(&params); err != nil {
			b.Error(err)
		}

		var sk *SecretKey
		var pk *PublicKey
		var err error

		kgen := bfvContext.NewKeyGenerator()

		// Public Key Generation
		b.Run(testString("KeyGen", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sk, pk = kgen.NewKeyPair()
				if err != nil {
					b.Error(err)
				}
			}
		})

		// Encryption
		encryptorPk, err := bfvContext.NewEncryptorFromPk(pk)
		encryptorSk, err := bfvContext.NewEncryptorFromSk(sk)

		if err != nil {
			b.Error(err)
		}

		ptcoeffs := bfvContext.NewRandomPlaintextCoeffs()
		pt := bfvContext.NewPlaintext()
		pt.setCoefficientsUint64(bfvContext, ptcoeffs)
		b.Run(testString("EncryptFromPkNew", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = encryptorPk.EncryptNew(pt)
			}
		})

		ctd1 := bfvContext.NewCiphertext(1)
		b.Run(testString("EncryptFromPk", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = encryptorPk.Encrypt(pt, ctd1)
			}
		})

		b.Run(testString("EncryptFromSkNew", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = encryptorSk.EncryptNew(pt)
			}
		})

		b.Run(testString("EncryptFromSk", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = encryptorSk.Encrypt(pt, ctd1)
			}
		})

		// Decryption
		decryptor, err := bfvContext.NewDecryptor(sk)
		if err != nil {
			b.Error(err)
		}
		ptp := bfvContext.NewPlaintext()
		b.Run(testString("Decrypt", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				decryptor.Decrypt(ctd1, ptp)
			}
			_ = ptp
		})

		evaluator := bfvContext.NewEvaluator()

		ct1, err := encryptorSk.EncryptNew(pt)
		if err != nil {
			b.Error(err)
		}

		ct2, err := encryptorSk.EncryptNew(pt)
		if err != nil {
			b.Error(err)
		}

		// Addition
		b.Run(testString("Add", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := evaluator.Add(ct1, ct2, ctd1); err != nil {
					b.Error(err)
				}
			}
		})

		// Subtraction
		b.Run(testString("Sub", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := evaluator.Sub(ct1, ct2, ctd1); err != nil {
					b.Error(err)
				}
			}
		})

		// Multiplication
		ctd2 := bfvContext.NewCiphertext(2)
		b.Run(testString("Multiply", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = evaluator.Mul(ct1, ct2, ctd2)
			}
		})

		// Square is Mul(ct, ct) for now
		b.Run(testString("Square", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = evaluator.Mul(ct1, ct1, ctd2)
			}
		})

		// Relinearization Key Generation not becnhmarked (no inplace gen)
		rlk := kgen.NewRelinKey(sk, 2)

		// Relinearization
		b.Run(testString("Relin", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := evaluator.Relinearize(ctd2, rlk, ctd1); err != nil {
					b.Error(err)
				}
			}
		})

		// Rotation Key Generation not benchmarked (no inplace gen)
		rtk := kgen.NewRotationKeysPow2(sk, true)

		// Rotation Rows
		b.Run(testString("RotateRows", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := evaluator.RotateRows(ct1, rtk, ctd1); err != nil {
					b.Error(err)
				}
			}
		})

		// Rotation Cols
		b.Run(testString("RotateCols", &params), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := evaluator.RotateColumns(ct1, 1, rtk, ctd1); err != nil {
					b.Error(err)
				}
			}
		})

	}
}

func testString(opname string, params *Parameters) string {
	return fmt.Sprintf("%s/params=%d", opname, params.N)
}
