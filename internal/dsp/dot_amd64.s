//go:build amd64 && !purego
// +build amd64,!purego

#include "textflag.h"

// Four independent scalar-order sums, evaluated together. MULPS and ADDPS
// deliberately remain separate: FMA would round differently from upstream.
TEXT ·Dot4(SB), NOSPLIT, $0-40
    MOVQ input+0(FP), AX
    MOVQ stride+8(FP), BX
    SHLQ $2, BX
    MOVQ weights+16(FP), CX
    MOVQ n+24(FP), DX
    MOVQ out+32(FP), DI
    XORPS X0, X0
    CMPQ DX, $2
    JL tail
loop:
    MOVSS (AX), X1
    MOVSS (AX)(BX*1), X3
    SHUFPS $0, X1, X1
    SHUFPS $0, X3, X3
    MOVUPS (CX), X2
    MOVUPS 16(CX), X4
    MULPS X1, X2
    MULPS X3, X4
    ADDPS X2, X0
    ADDPS X4, X0
    LEAQ (AX)(BX*2), AX
    ADDQ $32, CX
    SUBQ $2, DX
    CMPQ DX, $2
    JGE loop
tail:
    TESTQ DX, DX
    JLE done
    MOVSS (AX), X1
    SHUFPS $0, X1, X1
    MOVUPS (CX), X2
    MULPS X1, X2
    ADDPS X2, X0
done:
    MOVUPS X0, (DI)
    RET
