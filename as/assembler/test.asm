.org $8000
.byte "Hello, 6502!", 0

start:
   LDA #$00
   TAX
loop:
   LDA message,X
   BEQ done
   JSR $FFD2    ; CHROUT on Commodore 64
   INX
   JMP loop
done:
   RTS

message:
   .byte "Hello, World!", 0