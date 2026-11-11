//go:build amd64 && !purego
// +build amd64,!purego

#include "textflag.h"

// Preserve the sixteen terms of each sum; vectorize across output channels.
TEXT ·Window(SB), NOSPLIT, $0-32
    MOVQ input+0(FP), R8
    MOVQ pos+8(FP), R9
    SHLQ $2, R9
    MOVQ weights+16(FP), R10
    MOVQ out+24(FP), DI
    XORQ R11, R11
outer:
    LEAQ (R9)(R11*1), AX
    LEAQ (R10)(R11*1), CX
    MOVQ $8, DX
    XORPS X0, X0
inner:
    MOVQ AX, BX
    ANDQ $4095, BX
    MOVUPS (R8)(BX*1), X1
    MOVUPS (CX), X2
    MULPS X1, X2
    ADDPS X2, X0
    LEAQ 384(AX), BX
    ANDQ $4095, BX
    MOVUPS (R8)(BX*1), X1
    MOVUPS 128(CX), X2
    MULPS X1, X2
    ADDPS X2, X0
    ADDQ $512, AX
    ADDQ $256, CX
    DECQ DX
    JNZ inner
    MOVUPS X0, (DI)(R11*1)
    ADDQ $16, R11
    CMPQ R11, $128
    JL outer
    RET
