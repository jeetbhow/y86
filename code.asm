.pos 0x100
irmovq a, %r8
mrmovq 0(%r8), %r8
irmovq b, %r9
mrmovq 0(%r9), %r9
addq %r8, %r9
halt

.pos 0x1000
a: .quad 1
b: .quad 2
